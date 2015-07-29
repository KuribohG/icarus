package client

import (
	"errors"

	"github.com/applepi-icpc/icarus/task"
	"github.com/applepi-icpc/icarus/task/storage"
)

const (
	handleNameLengthLimit = 16

	ErrHandleNameTooLong = errors.New("Handle name too long")
	ErrWrongLogin        = errors.New("Wrong username or password")
)

// Client has 2 versions,
// one is in icarus and another one is in icarus-satellite.
// they should be distinguished from build flag.
// ("+build server" and "+build satellite")
//
// Server part invokes dispatcher to send task,
//   (most times the only needed work is
//    to use standard dispatcher functions)
// and the satellite part do the actual work.
type Client interface {
	// Factory functions of actual (not abstract) users and courses.
	MakeUser(userID string, password string) (task.User, error)
	MakeCourse(name string, desc string, token string) (task.Course, error)

	// List courses a user can elect.
	ListCourse(task.User) ([]storage.CourseData, error)
}

func RegisterHandle(handle string, cli Client) error {
	if len(handle) > handleNameLengthLimit {
		return ErrHandleNameTooLong
	}

	// TODO
}

func GetHandle(handle string) (Client, error) {
	// TODO
}
