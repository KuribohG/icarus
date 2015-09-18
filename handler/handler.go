package handler

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"

	"github.com/applepi-icpc/icarus"
	"github.com/applepi-icpc/icarus/client"
	"github.com/applepi-icpc/icarus/dispatcher/server"
	"github.com/applepi-icpc/icarus/task"
	"github.com/applepi-icpc/icarus/task/manager"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/thinxer/semikami"
	"golang.org/x/net/context"
)

type M map[string]interface{}

const SessionName = "icarus-session"

var store *sessions.CookieStore

var (
	flagCookieSecret = flag.String("cookie", "grimoire-of-alice", "Cookie secret")
	flagEdgeUser     = flag.String("id", "edge", "Username of superuser")
	flagEdgePassword = flag.String("pw", "password", "Password of superuser")
)

var (
	OK            = M{"okay": true}
	Unauthorized  = M{"error": "unauthorized"}
	Forbidden     = M{"error": "forbidden"}
	NotFound      = M{"error": "not found"}
	InternalError = M{"error": "internal error"}

	BadJSON       = M{"error": "failed to parse JSON"}
	UnknownHandle = M{"error": "unknown handle"}

	BadField = func(field string) M {
		return M{
			"error": fmt.Sprintf("failed to parse field `%s`", field),
		}
	}
	BadMake = func(element string) M {
		return M{
			"error": fmt.Sprintf("failed to make %s", element),
		}
	}
)

func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, code int, msg string) error {
	return WriteJSON(w, code, M{
		"error": msg,
	})
}

// Middlewares

type keyType int

const (
	keyRaw keyType = iota
	keyHandle
	keyHandleName
	keyUserData
	keyUser
	keyTaskID
	keyTask
	keyAllTaskData
	keyAllTask
)

func ParseJSON(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	m := make(map[string]json.RawMessage)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadJSON)
		return nil
	}
	return context.WithValue(ctx, keyRaw, m)
}
func GetJSONKey(ctx context.Context, key string) (json.RawMessage, bool) {
	m := ctx.Value(keyRaw).(map[string]json.RawMessage)
	v, ok := m[key]
	return v, ok
}
func GetJSONKeyAs(ctx context.Context, key string, v interface{}) error {
	raw, ok := GetJSONKey(ctx, key)
	if !ok {
		return errors.New(fmt.Sprintf("field %s does not exist", key))
	}
	return json.Unmarshal(raw, v)
}

// After `ParseJSON`
func ParseHandle(allowEmpty bool) kami.Middleware {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		var h string
		_, exist := GetJSONKey(ctx, "handle")
		if !exist {
			if allowEmpty {
				newCtx := context.WithValue(ctx, keyHandleName, "")
				return context.WithValue(newCtx, keyHandle, nil)
			} else {
				WriteJSON(w, http.StatusBadRequest, BadField("handle"))
				return nil
			}
		}

		err := GetJSONKeyAs(ctx, "handle", &h)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, BadField("handle"))
			return nil
		}
		cli, err := client.GetHandle(h)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, UnknownHandle)
			return nil
		}
		newCtx := context.WithValue(ctx, keyHandleName, h)
		return context.WithValue(newCtx, keyHandle, cli)
	}
}
func GetHandleName(ctx context.Context) string {
	return ctx.Value(keyHandleName).(string)
}
func GetHandle(ctx context.Context) client.Client {
	return ctx.Value(keyHandle).(client.Client)
}

// After `ParseHandle`
func ParseUser(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	var ud icarus.UserData
	err := GetJSONKeyAs(ctx, "user", &ud)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadField("user"))
		return nil
	}
	cli := GetHandle(ctx)
	u, err := client.MakeUserByData(cli, ud)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadMake("user"))
		return nil
	}
	newCtx := context.WithValue(ctx, keyUser, u)
	return context.WithValue(newCtx, keyUserData, ud)
}
func GetUser(ctx context.Context) icarus.User {
	return ctx.Value(keyUser).(icarus.User)
}
func GetUserData(ctx context.Context) icarus.UserData {
	return ctx.Value(keyUserData).(icarus.UserData)
}

func ParseTaskID(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	var ID int
	err := GetJSONKeyAs(ctx, "id", &ID)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadField("id"))
		return nil
	}

	// authenticate
	taskdata, err := manager.GetTaskData(ID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, NotFound)
		return nil
	}
	session, err := store.Get(r, SessionName)
	if err != nil {
		log.Errorf("Frontend: Failed to save session: %s", err.Error())
		WriteJSON(w, http.StatusInternalServerError, InternalError)
		return nil
	}
	userid, ok := session.Values["user"]
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, Unauthorized)
		return nil
	}
	if userid != *flagEdgeUser {
		handle, ok := session.Values["handle"]
		if !ok || userid != taskdata.User.UserID || handle != taskdata.Handle {
			WriteJSON(w, http.StatusForbidden, Forbidden)
			return nil
		}
	}

	// get real task
	t, err := manager.GetTask(ID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, NotFound)
		return nil
	}
	newCtx := context.WithValue(ctx, keyTaskID, ID)
	return context.WithValue(newCtx, keyTask, t)
}
func GetTask(ctx context.Context) *task.Task {
	return ctx.Value(keyTask).(*task.Task)
}
func GetTaskID(ctx context.Context) int {
	return ctx.Value(keyTaskID).(int)
}

func ParseAllTaskData(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	session, err := store.Get(r, SessionName)
	if err != nil {
		log.Errorf("Frontend: Failed to save session: %s", err.Error())
		WriteJSON(w, http.StatusInternalServerError, InternalError)
		return nil
	}
	userid, ok := session.Values["user"]
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, Unauthorized)
		return nil
	}
	handle, _ := session.Values["handle"]

	raw := manager.ListTasksData()
	if userid == *flagEdgeUser {
		return context.WithValue(ctx, keyAllTaskData, raw)
	} else {
		res := make([]icarus.TaskData, 0)
		for _, v := range raw {
			if v.User.UserID == userid && v.Handle == handle {
				res = append(res, v)
			}
		}
		return context.WithValue(ctx, keyAllTaskData, res)
	}
}
func GetAllTaskData(ctx context.Context) []icarus.TaskData {
	return ctx.Value(keyAllTaskData).([]icarus.TaskData)
}

func ParseAllTask(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	ctx = ParseAllTaskData(ctx, w, r)
	if ctx == nil {
		return nil
	}
	td := GetAllTaskData(ctx)
	res := make([]*task.Task, 0)
	for _, v := range td {
		t, err := manager.GetTask(v.ID)
		if err != nil {
			log.Warnf("Frontend: Task %d vanished when retriving.", v.ID)
			continue
		}
		res = append(res, t)
	}
	return context.WithValue(ctx, keyAllTask, res)
}
func GetAllTask(ctx context.Context) []*task.Task {
	return ctx.Value(keyAllTask).([]*task.Task)
}

func AuthEdgeUser(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	session, err := store.Get(r, SessionName)
	if err != nil {
		log.Errorf("Frontend: Failed to save session: %s", err.Error())
		WriteJSON(w, http.StatusInternalServerError, InternalError)
		return nil
	}
	userid, ok := session.Values["user"]
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, Unauthorized)
		return nil
	}
	if userid != *flagEdgeUser {
		WriteJSON(w, http.StatusForbidden, Forbidden)
		return nil
	}
	return ctx
}

// Do actual work

func InitHandler() {
	router := httprouter.New()
	amaterasu := kami.New(context.Background, router)
	store = sessions.NewCookieStore([]byte(*flagCookieSecret))

	amaterasu = amaterasu.With(ParseJSON)

	// Login
	// - Form: handle, userid, password
	//		(if userid == "edge", `handle` is not needed)
	// - Return: okay / error
	amaterasu.With(ParseHandle(true)).Post("/login", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var userid, password string
		err := GetJSONKeyAs(ctx, "userid", &userid)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, BadField("userid"))
			return
		}
		err = GetJSONKeyAs(ctx, "password", &password)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, BadField("password"))
			return
		}

		if userid == *flagEdgeUser {
			if password != *flagEdgePassword {
				WriteJSON(w, http.StatusForbidden, Forbidden)
				return
			}
		} else {
			cli := GetHandle(ctx)
			if cli == nil {
				WriteJSON(w, http.StatusBadRequest, BadField("handle"))
				return
			}
			u, err := cli.MakeUser(userid, password)
			if err != nil {
				WriteJSON(w, http.StatusBadRequest, BadMake("user"))
				return
			}
			_, err = u.Login()
			if err != nil {
				if err == server.ErrFailedToLogin {
					WriteJSON(w, http.StatusForbidden, Forbidden)
				} else {
					WriteJSON(w, http.StatusInternalServerError, InternalError)
				}
				return
			}
		}

		session, err := store.Get(r, SessionName)
		session.Values["user"] = userid
		session.Values["handle"] = GetHandleName(ctx)
		err = session.Save(r, w)
		if err != nil {
			log.Errorf("Frontend: Failed to save session: %s", err.Error())
			WriteJSON(w, http.StatusInternalServerError, InternalError)
		} else {
			WriteJSON(w, http.StatusOK, OK)
		}
	})

	// List courses
	// - Form: handle, User
	// - Return: []CourseData / error
	amaterasu.With(AuthEdgeUser).With(ParseHandle(false)).With(ParseUser).Post("/list_courses", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		u := GetUser(ctx)
		courses, err := u.ListCourse()
		if err != nil {
			if err == server.ErrInvalidData {
				WriteJSON(w, http.StatusInternalServerError, InternalError)
			} else {
				WriteError(w, http.StatusBadGateway, err.Error())
			}
			return
		}
		WriteJSON(w, http.StatusOK, courses)
	})

	// List tasks and statistics
	// - Return: []TaskData / error
	// -- omit User.Password
	// -- omit Course.Token
	amaterasu.With(ParseAllTaskData).Post("/list_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		td := GetAllTaskData(ctx)
		for k, _ := range td {
			for kk, _ := range td[k].Courses {
				td[k].Courses[kk].Token = ""
			}
			td[k].User.Password = ""
		}
		WriteJSON(w, http.StatusOK, td)
	})

	// Create task
	// - Form: handle, User, []Course
	// - Return: id / error
	amaterasu.With(ParseHandle(false)).With(ParseUser).Post("/create_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ud := GetUserData(ctx)

		var cd []icarus.CourseData
		err := GetJSONKeyAs(ctx, "courses", &cd)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, BadField("courses"))
			return
		}

		td := icarus.TaskData{
			Handle:  GetHandleName(ctx),
			User:    ud,
			Courses: cd,
		}
		id, t, err := manager.CreateTask(td)
		if err != nil {
			log.Warnf("Frontend: Error creating task: %s", err.Error())
			WriteJSON(w, http.StatusInternalServerError, InternalError)
			return
		}

		// Start task
		t.Start()

		WriteJSON(w, http.StatusOK, M{
			"id": id,
		})
	})

	// Start task
	// - Form: id
	// - Return: okay / error
	amaterasu.With(ParseTaskID).Post("/start_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetTask(ctx)
		t.Start()
		WriteJSON(w, http.StatusOK, OK)
	})

	// Stop task
	// ...
	amaterasu.With(ParseTaskID).Post("/stop_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetTask(ctx)
		t.Stop()
		WriteJSON(w, http.StatusOK, OK)
	})

	// Restart task
	// ...
	amaterasu.With(ParseTaskID).Post("/restart_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetTask(ctx)
		t.Restart()
		WriteJSON(w, http.StatusOK, OK)
	})

	// Delete task
	// ...
	amaterasu.With(ParseTaskID).Post("/delete_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ID := GetTaskID(ctx)
		err := manager.DeleteTask(ID)
		if err != nil {
			if err == manager.ErrNotFound {
				WriteJSON(w, http.StatusNotFound, NotFound)
			} else {
				log.Errorf("Frontend: Error deleting task: %s", err.Error())
				WriteJSON(w, http.StatusInternalServerError, InternalError)
			}
			return
		}
		WriteJSON(w, http.StatusOK, OK)
	})

	// Start all tasks
	// - Return: okay / error
	amaterasu.With(ParseAllTask).Post("/start_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetAllTask(ctx)
		for _, v := range t {
			v.Start()
		}
		WriteJSON(w, http.StatusOK, OK)
	})

	// Stop all tasks
	// ...
	amaterasu.With(ParseAllTask).Post("/stop_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetAllTask(ctx)
		for _, v := range t {
			v.Stop()
		}
		WriteJSON(w, http.StatusOK, OK)
	})

	// Restart all tasks
	// ...
	amaterasu.With(ParseAllTask).Post("/restart_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetAllTask(ctx)
		for _, v := range t {
			v.Restart()
		}
		WriteJSON(w, http.StatusOK, OK)
	})

	// Delete all tasks
	// ...
	amaterasu.With(ParseAllTaskData).Post("/delete_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t := GetAllTaskData(ctx)
		for _, v := range t {
			ID := v.ID
			err := manager.DeleteTask(ID)
			if err != nil {
				if err == manager.ErrNotFound {
					WriteJSON(w, http.StatusNotFound, NotFound)
				} else {
					log.Errorf("Frontend: Error deleting task: %s", err.Error())
					WriteJSON(w, http.StatusInternalServerError, InternalError)
				}
				return
			}
		}
		WriteJSON(w, http.StatusOK, OK)
	})

	http.Handle("/", amaterasu.Handler())
}
