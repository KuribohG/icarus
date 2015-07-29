package manager

import (
	"github.com/applepi-icpc/icarus/task"
	"github.com/applepi-icpc/icarus/task/storage"
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
	Header   storage.TaskData
	Instance task.Task
}

var tasks map[int]TaskEntry

func init() {
	tasks = make(map[int]TaskEntry)
}

func CreateTask(task TaskData) (int, error) {
	// TODO:
	// 1. Add to database (storage.CreateTask)
	// 2. Add to tasks
}

func DeleteTask(ID int) error {
	// TODO:
	// 1. Remove from data (storage.DeleteTask)
	// 2. Remove from tasks
}

func GetTask(ID int) (task.Task, error) {
	// TODO
}

func GetAllTasks() []task.Task {
	// TODO
}

func ListTask() []storage.TaskData {
	// TODO
}
