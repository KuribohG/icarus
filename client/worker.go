package client

import (
	"sync"
)

var workers map[string]Worker

// Worker is in icarus-satellite.
//
// What worker needs to do is to handle server's subtask and do actual response.
// All datum are transfered in []string so the worker need to understand server's
//   request correctly and generate suitable response. Errors should be coded in
//   response so worker cannot generate any Go-style errors.
type Worker interface {
	Login(data []string) []string
	ListCourse(data []string) []string
	Elect(data []string) []string
}

var workerIniter sync.Once

func initWorker() {
	workerIniter.Do(func() {
		workers = make(map[string]Worker)
	})
}

func RegisterWorker(handle string, w Worker) error {
	initWorker()
	if len(handle) > handleNameLengthLimit {
		return ErrHandleNameTooLong
	}

	_, ok := workers[handle]
	if ok {
		return ErrHandleExists
	}

	workers[handle] = w
	return nil
}

func RegisteredWorker() map[string]Worker {
	initWorker()
	res := make(map[string]Worker)
	for k, v := range workers {
		res[k] = v
	}
	return res
}

func RegisteredWorkerList() []string {
	initWorker()
	res := make([]string, 0)
	for k, _ := range workers {
		res = append(res, k)
	}
	return res
}

func GetWorker(handle string) (Worker, error) {
	initWorker()
	w, ok := workers[handle]
	if !ok {
		return nil, ErrHandleNotFound
	}
	return w, nil
}
