package storage

import (
	"database/sql"
	"flag"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/applepi-icpc/icarus"
	_ "github.com/go-sql-driver/mysql"
)

var (
	db     *sql.DB
	dbOnce sync.Once

	flagMySQL = flag.String("db", "root@tcp(127.0.0.1:3306)/icarus", "MySQL database for Icarus")
)

func InitStorage() {
	dbOnce.Do(func() {
		var err error
		db, err = sql.Open("mysql", *flagMySQL)
		if err != nil {
			panic(err)
		}
	})
}

func ListTasks() ([]icarus.TaskData, error) {
	InitStorage()
	rows, err := db.Query("SELECT task.id, task.handle, user.userid, user.password FROM task INNER JOIN user ON task.id = user.task_id")
	if err != nil {
		log.Warnf("storage.GetAllTasks(): Failed to fetch from database: %s.", err.Error())
		return nil, err
	}
	defer rows.Close()

	res := make([]icarus.TaskData, 0)
	for rows.Next() {
		var entry icarus.TaskData
		entry.Courses = make([]icarus.CourseData, 0)
		err = rows.Scan(&entry.ID, &entry.Handle, &entry.User.UserID, &entry.User.Password)
		if err != nil {
			log.Warnf("storage.GetAllTasks(): Failed to fetch a task: %s.", err.Error())
			continue
		}

		rowsCourse, err := db.Query("SELECT course.name, course.desc, course.token FROM course WHERE course.task_id = ?", entry.ID)
		if err != nil {
			log.Warnf("storage.GetAllTasks(): Task %d: Failed to fetch courses: %s.", entry.ID, err.Error())
			continue
		}
		func(r *sql.Rows) {
			defer r.Close()
			for r.Next() {
				var cs icarus.CourseData
				err = r.Scan(&cs.Name, &cs.Desc, &cs.Token)
				if err != nil {
					log.Warnf("storage.GetAllTasks(): Task %d: Failed to fetch a course: %s.", entry.ID, err.Error())
					continue
				}
				entry.Courses = append(entry.Courses, cs)
			}
		}(rowsCourse)

		res = append(res, entry)
	}
	return res, nil
}

// task.ID will be ignored.
func CreateTask(task icarus.TaskData) (ID int, err error) {
	InitStorage()
	var res sql.Result
	res, err = db.Exec("INSERT INTO task (handle) VALUES (?)", task.Handle)
	if err != nil {
		log.Warnf("storage.CreateTask(): Failed to insert task %v: %s.", task, err.Error())
		return
	}
	newID, err := res.LastInsertId()
	if err != nil {
		log.Warnf("storage.CreateTask(): Failed to get ID for task %v: %s.", task, err.Error())
	}
	ID = int(newID)

	_, err = db.Exec("INSERT INTO user (userid, password, task_id) VALUES (?, ?, ?)", task.User.UserID, task.User.Password, ID)
	if err != nil {
		log.Warnf("storage.CreateTask(): Failed to insert user for task %v: %s.", task, err.Error())
		return
	}

	for _, c := range task.Courses {
		_, err = db.Exec("INSERT INTO course (name, desc, token, task_id) VALUES (?, ?, ?, ?)", c.Name, c.Desc, c.Token, ID)
		if err != nil {
			log.Warnf("storage.CreateTask(): Failed to insert course %v for task %v: %s.", c, task, err.Error())
			return
		}
	}

	return
}

func DeleteTask(ID int) error {
	InitStorage()
	_, err := db.Exec("DELETE FROM task WHERE id = ?", ID)
	return err
}
