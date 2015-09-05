package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/applepi-icpc/icarus/client"

	log "github.com/Sirupsen/logrus"

	"github.com/applepi-icpc/icarus/dispatcher"
)

type M map[string]interface{}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	checkErr(err)
}

var PushTimeout = map[dispatcher.SubtaskType]time.Duration{
	dispatcher.SubtaskLogin: 10 * time.Second,
	dispatcher.SubtaskList:  30 * time.Second,
	dispatcher.SubtaskElect: 5 * time.Second,
}
var PullTimeout = 30 * time.Second

var (
	ErrTimeout = errors.New("subtask timeout")
)

type Dispatcher struct {
	mux *http.ServeMux

	mu  sync.RWMutex
	qmu sync.Mutex

	queueCapacity int

	queue  map[string]chan *dispatcher.Subtask      // Handler -> Queue
	result map[int64]chan *dispatcher.SubtaskResult // ID -> Result
	cipher map[int64][]byte                         // ID -> Cipher
}

func NewDispatcher(capacity int) *Dispatcher {
	t := &Dispatcher{
		mux: http.NewServeMux(),

		queueCapacity: capacity,

		queue:  make(map[string]chan *dispatcher.Subtask),
		result: make(map[int64]chan *dispatcher.SubtaskResult),
		cipher: make(map[int64][]byte),
	}

	t.mux.HandleFunc("/get_task", func(w http.ResponseWriter, r *http.Request) {
		var request dispatcher.TaskRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			log.Errorf("Dispatcher: error getting request: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// filter out unsupported handles
		registered := client.RegisteredHandle()
		rawAccepts := strings.Split(request.Accepts, ",")
		accepts := make([]string, 0)
		for _, v := range rawAccepts {
			_, ok := registered[v]
			if ok {
				accepts = append(accepts, v)
			}
		}

		subtask, err := t.PullSubtask(accepts)
		if err != nil {
			if err != ErrTimeout {
				log.Errorf("Dispatcher: error getting subtask: %s", err.Error())
				http.Error(w, "", http.StatusInternalServerError)
			} else {
				writeJSON(w, http.StatusOK, dispatcher.TaskResponse{
					OK: false,
				})
			}
			return
		}

		rawContent, err := json.Marshal(subtask)
		if err != nil {
			log.Errorf("Dispatcher: error marshalling subtask: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		content, err := Encrypt(string(rawContent), request.Cipher)
		if err != nil {
			log.Errorf("Dispatcher: error encrypting: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// record cipher
		t.mu.Lock()
		t.cipher[subtask.ID], _ = GetNakedKey(request.Cipher)
		t.mu.Unlock()

		writeJSON(w, http.StatusOK, dispatcher.TaskResponse{
			OK:      true,
			Content: content,
		})
	})

	t.mux.HandleFunc("/put_result", func(w http.ResponseWriter, r *http.Request) {
		var resp dispatcher.WorkResponse
		err := json.NewDecoder(r.Body).Decode(&resp)
		if err != nil {
			log.Errorf("Dispatcher: error getting work response: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		key, ok := t.cipher[resp.TaskID]
		if !ok {
			// HTTP 410 (Gone): Timeout, or no such task ID exists.
			http.Error(w, err.Error(), http.StatusGone)
			return
		}

		decrypted, err := Decrypt(resp.Content, key)
		if err != nil {
			log.Errorf("Dispatcher: error decrypting response: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var tres dispatcher.SubtaskResult
		err = json.Unmarshal([]byte(decrypted), &tres)
		if err != nil {
			log.Errorf("Dispatcher: error unmarshaling work response: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t.mu.RLock()
		ch, ok := t.result[resp.TaskID]
		if !ok {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}
		ch <- &tres
		t.mu.RUnlock()
	})

	return t
}

func (d *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func (d *Dispatcher) ensureQueue(c string) (queue chan *dispatcher.Subtask) {
	queue, ok := d.queue[c]
	if !ok {
		func() {
			d.qmu.Lock()
			defer d.qmu.Unlock()

			queue = make(chan *dispatcher.Subtask, d.queueCapacity)
			d.queue[c] = queue
		}()
	}
	return
}

func (d *Dispatcher) PushSubtask(s *dispatcher.Subtask) <-chan *dispatcher.SubtaskResult {
	d.mu.Lock()
	defer d.mu.Unlock()

	q := d.ensureQueue(s.Handler)
	timeout := PushTimeout[s.Type]
	res := make(chan *dispatcher.SubtaskResult)

	go func() {
		tc := time.After(timeout)

		select {
		case <-tc:
			res <- &dispatcher.SubtaskResult{
				Error: ErrTimeout,
			}
			return
		case q <- s:
			// do nothing
		}

		d.mu.Lock()
		ch := make(chan *dispatcher.SubtaskResult)
		d.result[s.ID] = ch
		d.mu.Unlock()

		select {
		case <-tc:
			res <- &dispatcher.SubtaskResult{
				Error: ErrTimeout,
			}
		case t := <-ch:
			res <- t
		}

		d.mu.Lock()
		delete(d.result, s.ID)
		close(ch)
		delete(d.cipher, s.ID)
		d.mu.Unlock()
	}()

	return res
}

func (d *Dispatcher) RunSubtask(s *dispatcher.Subtask) *dispatcher.SubtaskResult {
	ch := d.PushSubtask(s)
	return <-ch
}

func (d *Dispatcher) PullSubtask(accepts []string) (*dispatcher.Subtask, error) {
	tc := time.After(PullTimeout)
	cases := make([]reflect.SelectCase, len(accepts)+1)
	var timeoutIdx = len(accepts)
	for i, v := range accepts {
		q := d.ensureQueue(v)
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(q)}
	}
	cases[timeoutIdx] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(tc)}

	for {
		chosen, value, _ := reflect.Select(cases)
		if chosen == timeoutIdx {
			return nil, ErrTimeout
		}

		subtask := value.Interface().(*dispatcher.Subtask)
		d.mu.RLock()
		_, ok := d.result[subtask.ID]
		d.mu.RUnlock()
		if ok {
			return subtask, nil
		}
	}
}
