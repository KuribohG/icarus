package manager

import (
	"errors"
	"sync"

	"github.com/applepi-icpc/icarus"
	"github.com/applepi-icpc/icarus/client"
	"github.com/applepi-icpc/icarus/task"
	"github.com/applepi-icpc/icarus/task/storage"
)

type TaskEntry struct {
	Header   icarus.TaskData
	Instance *task.Task
}

var mu sync.Mutex
var tasks map[int]TaskEntry

var (
	ErrNotFound = errors.New("task ID not found")
)

func init() {
	tasks = make(map[int]TaskEntry)
}

// this function only generate a task and add it into `tasks`
func addTask(taskdata icarus.TaskData) (t *task.Task, err error) {
	c, err := client.GetHandle(taskdata.Handle)
	if err != nil {
		return
	}
	var user icarus.User
	user, err = client.MakeUserByData(c, taskdata.User)
	if err != nil {
		return
	}
	courses := make([]icarus.Course, 0)
	for _, csData := range taskdata.Courses {
		var cs icarus.Course
		cs, err = client.MakeCourceByData(c, csData)
		if err != nil {
			return
		}
		courses = append(courses, cs)
	}

	t = task.NewTask(user, courses)
	tasks[taskdata.ID] = TaskEntry{
		Header:   taskdata,
		Instance: t,
	}

	return
}

func CreateTask(taskdata icarus.TaskData) (ID int, t *task.Task, err error) {
	mu.Lock()
	defer mu.Unlock()

	ID, err = storage.CreateTask(taskdata)
	if err != nil {
		return
	}
	taskdata.ID = ID

	t, err = addTask(taskdata)
	return
}

// Once a task is deleted, it would be stopped first.
func DeleteTask(ID int) error {
	mu.Lock()
	defer mu.Unlock()

	entry, exist := tasks[ID]
	if !exist {
		return ErrNotFound
	}
	err := storage.DeleteTask(ID)
	if err != nil {
		return err
	}
	entry.Instance.Stop()
	delete(tasks, ID)
	return nil
}

func GetTask(ID int) (*task.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	entry, exist := tasks[ID]
	if !exist {
		return nil, ErrNotFound
	}
	return entry.Instance, nil
}

func ListTasks() []*task.Task {
	mu.Lock()
	defer mu.Unlock()

	res := make([]*task.Task, 0)
	for _, entry := range tasks {
		res = append(res, entry.Instance)
	}
	return res
}

func GetTaskData(ID int) (icarus.TaskData, error) {
	mu.Lock()
	defer mu.Unlock()

	entry, exist := tasks[ID]
	if !exist {
		return icarus.TaskData{}, ErrNotFound
	}
	return entry.Header, nil
}

func ListTasksData() []icarus.TaskData {
	mu.Lock()
	defer mu.Unlock()

	res := make([]icarus.TaskData, 0)
	for _, entry := range tasks {
		hdr := entry.Header
		hdr.Stat = entry.Instance.Statistics()
		res = append(res, hdr)
	}
	return res
}

// This function will panic on any error it encounters.
func InitManager() {
	tasksdata, err := storage.ListTasks()
	if err != nil {
		panic(err)
	}
	for _, t := range tasksdata {
		inst, err := addTask(t)

		// Start task
		inst.Start()

		if err != nil {
			panic(err)
		}
	}
}
