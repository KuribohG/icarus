package task

import (
	"errors"
	"log"
	"math/rand"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type testCourse struct {
	remaining   int32
	errorToMake int32
	elected     int32
}

const session = "xxx"

func NewTestCourse(remaining int32, errorToMake int32) testCourse {
	return testCourse{
		remaining:   remaining,
		errorToMake: errorToMake,
		elected:     0,
	}
}

func (t *testCourse) Name() string {
	return "Test Course"
}

func (t *testCourse) Description() string {
	return "Dummy courses for testing."
}

func (t *testCourse) Elect(k LoginSession) (bool, error) {
	log.Printf("Electing...Remaining %d, Elected %d", t.remaining, t.elected)
	if reflect.TypeOf(k).Name() != "string" {
		log.Fatalf("Wrong type of login session.")
	}
	if k.(string) != session {
		log.Fatalf("Wrong type of login session.")
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
			return true, nil
		}
	}
}

type testUser struct {
	errorToMake int32
	loginCount  int32
}

func (t *testUser) Name() string {
	return "Test User"
}

func (t *testUser) Login() (LoginSession, error) {
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

func TestBasic(t *testing.T) {
	MaxRetry = 5
	LoopInterval = 3 * time.Second

	var testCount int32 = 5

	user := testUser{}
	course := NewTestCourse(testCount, 0)
	testTask := NewTask(&user, []Course{&course})
	testTask.Start()
	time.Sleep(time.Duration(testCount+2) * LoopInterval)

	running, succeeded, failed, lastError, elected := testTask.Statistics()
	t.Logf("R = %t, S/F = %d/%d, LE = \"%s\", E = %t\n", running, succeeded, failed, lastError, elected)

	t.Logf("User: ER = %d, L = %d\n", user.errorToMake, user.loginCount)
	t.Logf("Course: R = %d, ER = %d, EL = %d\n", course.remaining, course.errorToMake, course.elected)

	if user.loginCount != 1 {
		t.Fatalf("Login for %d time(s).", user.loginCount)
	}
	if course.elected != testCount+1 {
		t.Fatalf("Elected for %d time(s).", course.elected)
	}
	if course.remaining != 0 {
		t.Fatalf("WTF...Remaining %d != 0.", course.remaining)
	}
	if running {
		t.Fatalf("Still running.")
	}
	if !elected {
		t.Fatalf("Should elected.")
	}
}
