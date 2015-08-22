package task

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus"
)

type Task struct {
	user    icarus.User
	courses []icarus.Course // A list concatenated by "or"

	mutex        sync.Mutex
	running      bool
	currentRunID int

	login   bool
	session icarus.LoginSession

	succeeded int64
	failed    int64
	lastError string
	elected   bool
}

// Make a new task.
func NewTask(user icarus.User, courses []icarus.Course) *Task {
	return &Task{
		user:    user,
		courses: courses,
	}
}

// Get this task's login user.
func (t *Task) User() icarus.User {
	return t.user
}

// Get this task's candidate courses.
func (t *Task) Courses() []icarus.Course {
	return t.courses
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
		session, err := t.user.Login()
		if err != nil {
			return false
		}
		t.login = true
		t.session = session
		return true
	} else {
		var wg sync.WaitGroup
		noError := true
		for _, v := range t.courses {
			wg.Add(1)

			// Elect a course
			go func(c icarus.Course) {
				defer wg.Done()
				elected, err := c.Elect(t.session)
				if err != nil {
					noError = false
					t.logError(err, fmt.Sprintf("%s: %s", t.user.Name(), c.Name()))
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

// Start this task.
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

// Terminate this task.
func (t *Task) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.running = false
	t.login = false
}

// Restart this task.
func (t *Task) Restart() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.running {
		t.login = false
	}
}

// Get this task's statistics.
func (t *Task) Statistics() icarus.Stat {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return icarus.Stat{t.running, t.succeeded, t.failed, t.lastError, t.elected}
}
