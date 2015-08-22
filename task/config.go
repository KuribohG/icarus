package task

import (
	"time"
)

var (
	MaxRetry     int           = 5
	LoopInterval time.Duration = 5 * time.Second
)
