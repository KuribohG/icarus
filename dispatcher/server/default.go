package server

import "github.com/applepi-icpc/icarus/client"

var DefaultDispatcher *Dispatcher

// It's main's obligation to invoke this function.
func InitDefaultDispatcher() {
	DefaultDispatcher = NewDispatcher(1024, client.RegisteredList())
}
