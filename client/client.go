package client

import (
	"errors"
	"sync"

	"github.com/applepi-icpc/icarus"
)

const (
	handleNameLengthLimit = 128
)

var registered map[string]Client

var (
	ErrHandleNameTooLong = errors.New("Handle name too long")
	ErrHandleNotFound    = errors.New("Handle not found")
	ErrHandleExists      = errors.New("Handle exists")
)

// Client is in icarus (server part).
//
// Server part invokes dispatcher to send task,
//   (most times the only needed work is
//    to use standard dispatcher functions)
// and the satellite part do the actual work.
type Client interface {
	// Factory functions of actual (not abstract) users and courses.
	MakeUser(userID string, password string) (icarus.User, error)
	MakeCourse(name string, desc string, token string) (icarus.Course, error)
}

var clientIniter sync.Once

func initInterface() {
	clientIniter.Do(func() {
		registered = make(map[string]Client)
	})
}

func MakeUserByData(c Client, data icarus.UserData) (icarus.User, error) {
	return c.MakeUser(data.UserID, data.Password)
}

func MakeCourceByData(c Client, data icarus.CourseData) (icarus.Course, error) {
	return c.MakeCourse(data.Name, data.Desc, data.Token)
}

func RegisterHandle(handle string, cli Client) error {
	initInterface()
	if len(handle) > handleNameLengthLimit {
		return ErrHandleNameTooLong
	}

	_, ok := registered[handle]
	if ok {
		return ErrHandleExists
	}

	registered[handle] = cli
	return nil
}

func RegisteredHandle() map[string]Client {
	initInterface()
	res := make(map[string]Client)
	for k, v := range registered {
		res[k] = v
	}
	return res
}

func RegisteredList() []string {
	initInterface()
	res := make([]string, 0)
	for k, _ := range registered {
		res = append(res, k)
	}
	return res
}

func GetHandle(handle string) (Client, error) {
	initInterface()
	cli, ok := registered[handle]
	if !ok {
		return nil, ErrHandleNotFound
	}
	return cli, nil
}
