package satellite

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/applepi-icpc/icarus/dispatcher"
)

type PostOffice struct {
	root              string
	avaliableHandlers []string
}

func NewPostOffice(root string, avaliableHandlers []string) *PostOffice {
	return &PostOffice{
		root:              root,
		avaliableHandlers: avaliableHandlers,
	}
}

// Handles a specific subtask
type Postman struct {
	office  *PostOffice
	key     []byte
	Subtask *dispatcher.Subtask
}

var (
	ErrNoNewTasks   = errors.New("no new tasks")
	ErrTaskVanished = errors.New("task vanished")
)

func (p *PostOffice) GetTask() (*Postman, error) {
	orig, cipher := GenKey()
	request := dispatcher.TaskRequest{
		Accepts: strings.Join(p.avaliableHandlers, ","),
		Cipher:  cipher,
	}
	requestBody, err := json.Marshal(request)
	checkErr(err)

	resp, err := http.Post(fmt.Sprintf("%s/get_task", p.root),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		log.Warnf("GetTask: Failed to make get task request: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Warnf("GetTask: Server: HTTP %d: %s", resp.StatusCode, string(b))
		return nil, errors.New(string(b))
	}

	var response dispatcher.TaskResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Warnf("GetTask: Failed to decode server's response: %s", err.Error())
		return nil, err
	}

	if !response.OK {
		return nil, ErrNoNewTasks
	}

	content, err := Decrypt(response.Content, orig)
	if err != nil {
		log.Warnf("GetTask: Failed to decrypt server's response: %s", err.Error())
		return nil, err
	}
	var sb dispatcher.Subtask
	err = json.Unmarshal([]byte(content), &sb)
	if err != nil {
		log.Warnf("GetTask: Failed to decode server's content: %s", err.Error())
		return nil, err
	}

	return &Postman{
		office:  p,
		key:     orig,
		Subtask: &sb,
	}, nil
}

func (pm *Postman) SendResult(res *dispatcher.SubtaskResult) error {
	rawContent, err := json.Marshal(res)
	if err != nil {
		return err
	}
	content, err := NakedEncrypt(string(rawContent), pm.key)
	if err != nil {
		return err
	}

	wr := dispatcher.WorkResponse{
		Content: content,
		TaskID:  pm.Subtask.ID,
	}
	requestBody, err := json.Marshal(wr)
	checkErr(err)

	resp, err := http.Post(fmt.Sprintf("%s/put_result", pm.office.root),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		log.Warnf("SendResult: Failed to make send result request: %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		log.Warnf("SendResult: Task vanished.")
		return ErrTaskVanished
	} else if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Warnf("SendResult: Server: HTTP %d: %s", resp.StatusCode, string(b))
		return errors.New(string(b))
	} else {
		return nil
	}
}
