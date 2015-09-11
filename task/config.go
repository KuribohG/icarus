package task

import (
	"errors"
	"time"
)

var (
	MaxRetry     int           = 5
	LoopInterval time.Duration = 5 * time.Second
)

var (
	// If this error was returned by Course.Elect, the task would be restarted immediately.
	ErrSessionExpired = errors.New("icarus: session expired")
)
