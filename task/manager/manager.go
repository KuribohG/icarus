package manager

import (
	"github.com/applepi-icpc/icarus"
	"github.com/applepi-icpc/icarus/task"
)

type Stat struct {
	Running   bool   `json:"running"`
	Succeeded int64  `json:"succeeded"`
	Failed    int64  `json:"failed"`
	LastError string `json:"last_error"`
	Elected   bool   `json:"elected"`
}

func InitManager() {
	// TODO:
	// 1. Fetch from database (storage.GetAllTasks)
	// 2. All 'em all to tasks
}

type TaskEntry struct {
	Header   icarus.TaskData
	Instance task.Task
}

var tasks map[int]TaskEntry

func init() {
	tasks = make(map[int]TaskEntry)
}

func CreateTask(task icarus.TaskData) (int, error) {
	// TODO:
	// 1. Add to database (storage.CreateTask)
	// 2. Add to tasks
	return 0, nil
}

func DeleteTask(ID int) error {
	// TODO:
	// 1. Remove from data (storage.DeleteTask)
	// 2. Remove from tasks
	return nil
}

func GetTask(ID int) (*task.Task, error) {
	// TODO
	return nil, nil
}

func GetAllTasks() []*task.Task {
	// TODO
	return nil
}

func ListTask() []icarus.TaskData {
	// TODO
	return nil
}
