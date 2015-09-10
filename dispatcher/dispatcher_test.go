package dispatcher_test

import (
	"flag"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus/dispatcher"
	"github.com/applepi-icpc/icarus/dispatcher/satellite"
	"github.com/applepi-icpc/icarus/dispatcher/server"
)

var (
	flagTaskBind = flag.String("task", ":8001", "Task bind address")
	flagRoot     = flag.String("root", "http://127.0.0.1:8001", "API Root")
)

func testSatellite(name string, sep string, delay time.Duration, t *testing.T) {
	p := satellite.NewPostOffice(*flagRoot, []string{name})
	for i := 0; i < 6; i++ {
		pm, err := p.GetTask()
		if err != nil {
			if err == satellite.ErrNoNewTasks {
				log.Printf("Satellite %s: no new tasks", name)
			} else {
				t.Fatalf("Error fetching new task: %s", err.Error())
			}
			continue
		}

		sb := pm.Subtask
		if sb.Handler != name {
			t.Fatalf("Satellite %s: Fetched wrong subtasks.", name)
		}

		log.Printf("Satellite %s: get a task with ID %d.", name, sb.ID)

		resp := &dispatcher.SubtaskResult{
			Data: []string{strings.Join(sb.Data, sep)},
		}

		time.Sleep(delay)
		err = pm.SendResult(resp)
		if err == satellite.ErrTaskVanished {
			log.Warnf("Satellite %s: Task %d has gone!", name, sb.ID)
		} else if err != nil {
			log.Warnf("Satellite %s: Task %d: %s", name, sb.ID, err.Error())
		}
	}
}

var disp = server.NewDispatcher(1024, []string{"comma", "dot", "space"})
var once sync.Once

func initDispatcher() {
	once.Do(func() {
		go func() {
			log.Printf("Task Handler at %s", *flagTaskBind)
			mux := http.NewServeMux()
			mux.Handle("/", disp)
			panic(http.ListenAndServe(*flagTaskBind, mux))
		}()
		time.Sleep(time.Second * 2)
	})
}

func TestDispatcherBasic(t *testing.T) {
	initDispatcher()

	server.PullTimeout = time.Second * 4

	go testSatellite("comma", ",", 0, t)
	go testSatellite("dot", ".", 0, t)
	go testSatellite("space", " ", 0, t)

	for i := 0; i < 3; i++ {
		go func(i int) {
			time.Sleep(time.Second * time.Duration(i*5))
			log.Printf("Send subtask comma %d", i)
			res := disp.RunSubtask(&dispatcher.Subtask{
				Handler: "comma",
				Type:    dispatcher.SubtaskElect,
				Data:    []string{"marisa", "alice"},
			})
			if res.Error != nil {
				log.Warnf("Task comma %d: %s", i, res.Error.Error())
			} else {
				if len(res.Data) != 1 || res.Data[0] != "marisa,alice" {
					t.Fatalf("Task comma %d: Wrong answer: %v", i, res.Data)
				} else {
					log.Printf("Task comma %d: Get result %v", i, res.Data)
				}
			}
		}(i)
	}

	for i := 0; i < 3; i++ {
		go func(i int) {
			time.Sleep(time.Second * time.Duration(i*5))
			log.Printf("Send subtask dot %d", i)
			res := disp.RunSubtask(&dispatcher.Subtask{
				Handler: "dot",
				Type:    dispatcher.SubtaskElect,
				Data:    []string{"marisa", "alice"},
			})
			if res.Error != nil {
				log.Warnf("Task dot %d: %s", i, res.Error.Error())
			} else {
				if len(res.Data) != 1 || res.Data[0] != "marisa.alice" {
					t.Fatalf("Task dot %d: Wrong answer: %v", i, res.Data)
				} else {
					log.Printf("Task dot %d: Get result %v", i, res.Data)
				}
			}
		}(i)
	}

	for i := 0; i < 3; i++ {
		go func(i int) {
			time.Sleep(time.Second * time.Duration(i*5))
			log.Printf("Send subtask space %d", i)
			res := disp.RunSubtask(&dispatcher.Subtask{
				Handler: "space",
				Type:    dispatcher.SubtaskElect,
				Data:    []string{"marisa", "alice"},
			})
			if res.Error != nil {
				log.Warnf("Task space %d: %s", i, res.Error.Error())
			} else {
				if len(res.Data) != 1 || res.Data[0] != "marisa alice" {
					t.Fatalf("Task space %d: Wrong answer: %v", i, res.Data)
				} else {
					log.Printf("Task space %d: Get result %v", i, res.Data)
				}
			}
		}(i)
	}

	time.Sleep(time.Second * 21)
}

func TestDispatcherBasic2(t *testing.T) {
	initDispatcher()

	server.PullTimeout = time.Second * 4

	go testSatellite("comma", ",", 0, t)
	go testSatellite("comma", ",", 0, t)
	go testSatellite("comma", ",", 0, t)

	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			go func(i int) {
				time.Sleep(time.Second * time.Duration(i*5))
				log.Printf("Send subtask comma %d", i)
				res := disp.RunSubtask(&dispatcher.Subtask{
					Handler: "comma",
					Type:    dispatcher.SubtaskElect,
					Data:    []string{"marisa", "alice"},
				})
				if res.Error != nil {
					log.Warnf("Task comma %d: %s", i, res.Error.Error())
				} else {
					if len(res.Data) != 1 || res.Data[0] != "marisa,alice" {
						t.Fatalf("Task comma %d: Wrong answer: %v", i, res.Data)
					} else {
						log.Printf("Task comma %d: Get result %v", i, res.Data)
					}
				}
			}(i)
		}
	}

	time.Sleep(time.Second * 21)
}

func TestDispatcherTimeout(t *testing.T) {
	initDispatcher()

	server.PullTimeout = time.Second * 1

	go testSatellite("comma", ",", 0, t)
	go testSatellite("comma", ",", 0, t)
	go testSatellite("comma", ",", 0, t)

	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			go func(i int) {
				time.Sleep(time.Second * time.Duration(i*5))
				log.Printf("Send subtask comma %d", i)
				res := disp.RunSubtask(&dispatcher.Subtask{
					Handler: "comma",
					Type:    dispatcher.SubtaskElect,
					Data:    []string{"marisa", "alice"},
				})
				if res.Error != nil {
					log.Warnf("Task comma %d: %s", i, res.Error.Error())
				} else {
					if len(res.Data) != 1 || res.Data[0] != "marisa,alice" {
						t.Fatalf("Task comma %d: Wrong answer: %v", i, res.Data)
					} else {
						log.Printf("Task comma %d: Get result %v", i, res.Data)
					}
				}
			}(i)
		}
	}

	time.Sleep(time.Second * 21)
}

func TestDispatcherGone(t *testing.T) {
	initDispatcher()

	server.PullTimeout = time.Second * 4
	server.PushTimeout[dispatcher.SubtaskElect] = time.Second * 2

	go testSatellite("comma", ",", time.Second*3, t)
	go testSatellite("comma", ",", time.Second*3, t)
	go testSatellite("comma", ",", time.Second*3, t)

	for j := 0; j < 3; j++ {
		for i := 0; i < 3; i++ {
			go func(i int) {
				time.Sleep(time.Second * time.Duration(i*5))
				log.Printf("Send subtask comma %d", i)
				res := disp.RunSubtask(&dispatcher.Subtask{
					Handler: "comma",
					Type:    dispatcher.SubtaskElect,
					Data:    []string{"marisa", "alice"},
				})
				if res.Error != nil {
					log.Warnf("Task comma %d: %s", i, res.Error.Error())
				} else {
					if len(res.Data) != 1 || res.Data[0] != "marisa,alice" {
						t.Fatalf("Task comma %d: Wrong answer: %v", i, res.Data)
					} else {
						log.Printf("Task comma %d: Get result %v", i, res.Data)
					}
				}
			}(i)
		}
	}

	time.Sleep(time.Second * 21)
}
