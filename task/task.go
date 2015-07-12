package task

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type LoginSession interface{}

type Course interface {
	Name() string
	Description() string
	Elect(LoginSession) (bool, error)
}

type User interface {
	Name() string
	Login() (LoginSession, error)
}

type Task struct {
	User    User
	Courses []Course // A list concatenated by "or"

	mutex   sync.Mutex
	statex  sync.RWMutex
	signal  chan bool
	running bool

	login   bool
	session LoginSession

	succeeded int64 // Number of succeeded attempt
	failed    int64 // Number of failed attempt
	lastError string
	elected   bool
}

func NewTask(user User, courses []Course) *Task {
	return &Task{
		User:    user,
		Courses: courses,
		signal:  make(chan bool),
	}
}

func (t *Task) logError(err error, msg string) {
	t.statex.Lock()
	defer t.statex.Unlock()

	log.Warnf("%s: %s", msg, err.Error())
	t.failed++
	t.lastError = err.Error()
}

func (t *Task) logOK() {
	t.statex.Lock()
	defer t.statex.Unlock()

	t.succeeded++
}

func (t *Task) runOnce() bool {
	if !t.login {
		session, err := t.User.Login()
		if err != nil {
			return false
		}
		t.login = true
		t.session = session
		return true
	} else {
		var wg sync.WaitGroup
		noError := true
		for _, v := range t.Courses {
			wg.Add(1)

			// Elect a course
			go func(c Course) {
				defer wg.Done()
				elected, err := c.Elect(t.session)
				if err != nil {
					noError = false
					t.logError(err, fmt.Sprintf("%s: %s", t.User.Name(), c.Name()))
					return
				}

				t.logOK()

				if elected {
					t.elected = true
					go t.Stop()
				}
			}(v)
		}
		wg.Wait()
		if t.elected {
			return true
		} else {
			return noError
		}
	}
}

func (t *Task) run(signal chan bool) {
	retried := 0
	for {
		terminated := false
		endOfTurn := time.After(LoopInterval)

		select {
		case <-signal:
			// log.Println("terminating 1")
			terminated = true
		default:
			// log.Println("continue")
			func() {
				t.mutex.Lock()
				defer t.mutex.Unlock()

				// Do major work
				ok := t.runOnce()

				select {
				case <-signal:
					// log.Println("terminating 2")
					terminated = true
				default:
					if !ok {
						retried++
						if retried >= MaxRetry {
							go t.Restart()
						}
					} else {
						retried = 0
					}
				}
			}()
		}

		if terminated {
			break
		}

		<-endOfTurn
	}
	// log.Println("out of loop")
}

func (t *Task) Start() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.statex.Lock()
	defer t.statex.Unlock()

	if !t.running {
		t.running = true
		t.elected = false

		go t.run(t.signal)
	}
}

func (t *Task) Stop() {
	// log.Println("stop")
	t.signal <- true

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.statex.Lock()
	defer t.statex.Unlock()

	// Send signal to terminate the old loop
	t.signal = make(chan bool)
	t.running = false

	t.login = false

	t.succeeded = 0
	t.failed = 0
	t.lastError = ""
}

func (t *Task) Restart() {
	t.Stop()
	t.Start()
}

func (t *Task) Statistics() (running bool, succeeded int64, failed int64, lastError string, elected bool) {
	t.statex.RLock()
	defer t.statex.RUnlock()

	return t.running, t.succeeded, t.failed, t.lastError, t.elected
}
