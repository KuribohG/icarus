package storage

import (
	"database/sql"
	"flag"
	"sync"

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

func GetAllTasks() ([]icarus.TaskData, error) {
	InitStorage()
	rows, err := db.Query("SELECT task.id, task.handle, user.userid, user.password FROM task INNER JOIN user ON task.id = user.task_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// res := make([]icarus.TaskData, 0)
	for rows.Next() {
		var entry icarus.TaskData
		err = rows.Scan(&entry.ID, &entry.Handle, &entry.User.UserID, &entry.User.Password)
	}
	return nil, nil
}

// task.ID will be ignored.
func CreateTask(task icarus.TaskData) (int, error) {
	// TODO
	return 0, nil
}

func DeleteTask(ID int) error {
	// TODO
	return nil
}
