package storage

import (
	"database/sql"
	"flag"
	"sync"

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

type UserData struct {
	UserID   string `json:"userid"`
	Password string `json:"password"`
}
type CourseData struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Token string `json:"token"`
}
type TaskData struct {
	ID      int      `json:"id"`
	Handle  string   `json:"handle"`
	User    UserID   `json:"user"`
	Courses []Course `json:"courses"`
}

func GetAllTasks() ([]TaskData, error) {
	// TODO
}

// task.ID will be ignored.
func CreateTask(task TaskData) (int, error) {
	// TODO
}

func DeleteTask(ID int) error {
	// TODO
}
