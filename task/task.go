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

	mutex        sync.Mutex
	running      bool
	currentRunID int

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
	}
}

func (t *Task) logError(err error, msg string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Warnf("%s: %s", msg, err.Error())
	t.failed++
	t.lastError = err.Error()
}

func (t *Task) logOK() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.succeeded++
}

func (t *Task) logElected() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.elected = true
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
					t.logElected()
					t.Stop()
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

func (t *Task) run(runID int) {
	retried := 0
	for t.running && runID == t.currentRunID {
		endOfTurn := time.After(LoopInterval)

		// Do major work
		ok := t.runOnce()

		func() {
			t.mutex.Lock()
			defer t.mutex.Unlock()

			if !ok {
				retried++
				if retried >= MaxRetry {
					// Restart
					t.login = false
				}
			} else {
				retried = 0
			}
		}()

		<-endOfTurn
	}
}

func (t *Task) Start() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.running {
		t.running = true
		t.login = false
		t.elected = false

		t.currentRunID++
		go t.run(t.currentRunID)
	}
}

func (t *Task) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.running = false
	t.login = false
}

func (t *Task) Restart() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.running {
		t.login = false
	}
}

func (t *Task) Statistics() (running bool, succeeded int64, failed int64, lastError string, elected bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.running, t.succeeded, t.failed, t.lastError, t.elected
}
