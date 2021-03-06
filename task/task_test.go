package task

import (
	"errors"
	"log"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/applepi-icpc/icarus"
)

type testCourse struct {
	remaining    int32
	errorToMake  int32
	elected      int32
	concurrent   int32
	noConcurrent bool
	forbidden    bool
}

const session = "xxx"

func NewTestCourse(remaining int32, errorToMake int32) testCourse {
	return testCourse{
		remaining:   remaining,
		errorToMake: errorToMake,
		concurrent:  0,
		elected:     0,
	}
}

func (t *testCourse) Name() string {
	return "Test Course"
}

func (t *testCourse) Elect(k icarus.LoginSession) (bool, error) {
	conc := atomic.AddInt32(&t.concurrent, 1)
	if conc > 1 && t.noConcurrent {
		log.Fatalf("2 goroutines electing at the same time!")
	}
	defer atomic.AddInt32(&t.concurrent, -1)

	if t.forbidden {
		log.Fatalf("Entered forbidden elect function.")
	}

	log.Printf("Electing...Remaining %d, Elected %d", t.remaining, t.elected)
	if reflect.TypeOf(k).Name() != "string" {
		log.Fatalf("Wrong type of login session.")
	}
	if k.(string) != session {
		log.Fatalf("Wrong value of login session.")
	}

	atomic.AddInt32(&t.elected, 1)
	time.Sleep(time.Duration(rand.Float32()*3000) * time.Millisecond)
	if atomic.LoadInt32(&t.errorToMake) > 0 {
		atomic.AddInt32(&t.errorToMake, -1)
		return false, errors.New("False Alarm")
	} else {
		if atomic.LoadInt32(&t.remaining) > 0 {
			atomic.AddInt32(&t.remaining, -1)
			return false, nil
		} else {
			log.Println("Elected!")
			return true, nil
		}
	}
}

type testUser struct {
	errorToMake  int32
	loginCount   int32
	concurrent   int32
	noConcurrent bool
	forbidden    bool
}

func (t *testUser) Name() string {
	return "Test User"
}

func (t *testUser) Login() (icarus.LoginSession, error) {
	conc := atomic.AddInt32(&t.concurrent, 1)
	if conc > 1 && t.noConcurrent {
		log.Fatalf("2 goroutines electing at the same time!")
	}
	defer atomic.AddInt32(&t.concurrent, -1)

	if t.forbidden {
		log.Fatalf("Entered forbidden login function.")
	}

	log.Printf("Login...")
	atomic.AddInt32(&t.loginCount, 1)
	time.Sleep(time.Duration(rand.Float32()*3000) * time.Millisecond)
	if atomic.LoadInt32(&t.errorToMake) > 0 {
		atomic.AddInt32(&t.errorToMake, -1)
		return nil, errors.New("False Alarm")
	} else {
		return session, nil
	}
}

func (t *testUser) ListCourse() ([]icarus.CourseData, error) {
	return nil, nil
}

func TestBasic(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second
	log.Printf("TestBasic: Elect every %.2f seconds.", float32(LoopInterval/time.Second))

	var testCount int32 = 5

	user := testUser{}
	course := NewTestCourse(testCount, 0)
	testTask := NewTask(&user, []icarus.Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(testCount+2) * LoopInterval)

	s := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", s.Running, s.Succeeded, s.Failed, s.LastError, s.Elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if user.loginCount > 1 {
		t.Fatalf("Login for %d time(s).", user.loginCount)
	}
	if course.elected != testCount+1 {
		t.Fatalf("Elected for %d time(s).", course.elected)
	}
	if course.remaining != 0 {
		t.Fatalf("WTF...Remaining %d != 0.", course.remaining)
	}
	if s.Running {
		t.Fatalf("Still running.")
	}
	if !s.Elected {
		t.Fatalf("Should elected.")
	}
}

func TestStartStop(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second
	log.Printf("TestStartStop: Elect every %.2f seconds.", float32(LoopInterval/time.Second))

	var testCount int32 = 5

	user := testUser{}
	course := NewTestCourse(testCount, 0)
	testTask := NewTask(&user, []icarus.Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(rand.Float32()*100) * time.Millisecond)
	testTask.Stop()
	time.Sleep(time.Duration(rand.Float32()*100) * time.Millisecond)
	testTask.Start()
	time.Sleep(time.Duration(2) * LoopInterval)
	course.noConcurrent = true
	user.noConcurrent = true
	time.Sleep(time.Duration(testCount+2) * LoopInterval)

	s := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", s.Running, s.Succeeded, s.Failed, s.LastError, s.Elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if user.loginCount > 2 {
		t.Fatalf("Login for %d time(s).", user.loginCount)
	}
	if course.elected != testCount+1 {
		t.Fatalf("Elected for %d time(s).", course.elected)
	}
	if course.remaining != 0 {
		t.Fatalf("WTF...Remaining %d != 0.", course.remaining)
	}
	if s.Running {
		t.Fatalf("Still running.")
	}
	if !s.Elected {
		t.Fatalf("Should elected.")
	}
}

func TestRestart(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second
	log.Printf("TestRestart: Elect every %.2f seconds.", float32(LoopInterval/time.Second))

	var testCount int32 = 5
	var errorToMake int32 = 10

	user := testUser{}
	course := NewTestCourse(testCount, 0)
	course.errorToMake = errorToMake
	testTask := NewTask(&user, []icarus.Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(testCount+errorToMake+5) * LoopInterval)

	s := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", s.Running, s.Succeeded, s.Failed, s.LastError, s.Elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if user.loginCount > 3 {
		t.Fatalf("Login for %d time(s).", user.loginCount)
	}
	if course.elected != testCount+errorToMake+1 {
		t.Fatalf("Elected for %d time(s).", course.elected)
	}
	if course.remaining != 0 {
		t.Fatalf("WTF...Remaining %d != 0.", course.remaining)
	}
	if s.Running {
		t.Fatalf("Still running.")
	}
	if !s.Elected {
		t.Fatalf("Should elected.")
	}
}

func TestStop(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second
	log.Printf("TestStop: Elect every %.2f seconds.", float32(LoopInterval/time.Second))

	var testCount int32 = 5
	user := testUser{}
	course := NewTestCourse(testCount, 0)
	testTask := NewTask(&user, []icarus.Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(2) * LoopInterval)

	log.Println("Stop!")
	testTask.Stop()
	course.forbidden = true
	user.forbidden = true
	time.Sleep(time.Duration(2) * LoopInterval)

	s := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", s.Running, s.Succeeded, s.Failed, s.LastError, s.Elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if s.Running {
		t.Fatalf("Still running.")
	}
	if s.Elected {
		t.Fatalf("Should not elected.")
	}
}

func TestStopWhenRestarting(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second
	log.Printf("TestStopWhenRestarting: Elect every %.2f seconds.", float32(LoopInterval/time.Second))

	var testCount int32 = 5
	var errorToMake int32 = 5

	user := testUser{}
	course := NewTestCourse(testCount, 0)
	course.errorToMake = errorToMake
	testTask := NewTask(&user, []icarus.Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(MaxRetry+1)*LoopInterval + 1)
	log.Println("Stop!")
	testTask.Stop()
	course.forbidden = true
	user.forbidden = true
	time.Sleep(time.Duration(2) * LoopInterval)

	s := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", s.Running, s.Succeeded, s.Failed, s.LastError, s.Elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if user.loginCount > 1 {
		t.Fatalf("Login for %d time(s).", user.loginCount)
	}
	if s.Running {
		t.Fatalf("Still running.")
	}
	if s.Elected {
		t.Fatalf("Should not elected.")
	}
}
