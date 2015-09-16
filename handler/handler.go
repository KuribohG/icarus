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

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/thinxer/semikami"
	"golang.org/x/net/context"
)

type M map[string]interface{}

const SessionName = "icarus-session"

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

	BadJSON        = M{"error": "failed to parse JSON"}
	UnknownHandler = M{"error": "unknown handler"}

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
	keyHandler
	keyHandlerName
	keyUser
	keyTask
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
func ParseHandler(allowEmpty bool) kami.Middleware {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		var h string
		_, exist := GetJSONKey(ctx, "handler")
		if !exist {
			if allowEmpty {
				newCtx := context.WithValue(ctx, keyHandlerName, "")
				return context.WithValue(newCtx, keyHandler, nil)
			} else {
				WriteJSON(w, http.StatusBadRequest, BadField("handler"))
				return nil
			}
		}

		err := GetJSONKeyAs(ctx, "handler", &h)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, BadField("handler"))
			return nil
		}
		cli, err := client.GetHandle(h)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, UnknownHandler)
			return nil
		}
		newCtx := context.WithValue(ctx, keyHandlerName, h)
		return context.WithValue(newCtx, keyHandler, cli)
	}
}
func GetHandlerName(ctx context.Context) string {
	return ctx.Value(keyHandlerName).(string)
}
func GetHandler(ctx context.Context) client.Client {
	return ctx.Value(keyHandler).(client.Client)
}

// After `ParseHandler`
func ParseUser(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	var ud icarus.UserData
	err := GetJSONKeyAs(ctx, "user", &ud)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadField("user"))
		return nil
	}
	cli := GetHandler(ctx)
	u, err := client.MakeUserByData(cli, ud)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, BadMake("user"))
		return nil
	}
	return context.WithValue(ctx, keyUser, u)
}
func GetUser(ctx context.Context) icarus.User {
	return ctx.Value(keyUser).(icarus.User)
}

func ParseTaskID(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	// UNC
	return nil
}
func GetTask(ctx context.Context) *task.Task {
	return ctx.Value(keyTask).(*task.Task)
}

func InitHandler() {
	router := httprouter.New()
	amaterasu := kami.New(context.Background, router)
	store := sessions.NewCookieStore([]byte(*flagCookieSecret))

	amaterasu = amaterasu.With(ParseJSON)

	// Login
	// - Form: handler, userid, password
	//		(if userid == "edge", `handler` is not needed)
	// - Return: okay / error
	amaterasu.With(ParseHandler(true)).Post("/login", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
			cli := GetHandler(ctx)
			if cli == nil {
				WriteJSON(w, http.StatusBadRequest, BadField("handler"))
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
		session.Values["handler"] = GetHandlerName(ctx)
		err = session.Save(r, w)
		if err != nil {
			log.Errorf("Failed to save session: %s", err.Error())
			WriteJSON(w, http.StatusInternalServerError, InternalError)
		} else {
			WriteJSON(w, http.StatusOK, OK)
		}
	})

	// List courses
	// - Form: handler, User
	// - Return: []CourseData / error
	amaterasu.With(ParseHandler(false)).With(ParseUser).Get("/list_course", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// List tasks and statistics
	// - Return: []TaskData / error
	// -- omit User.Password
	// -- omit Course.Token
	amaterasu.Get("/list_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Create task
	// - Form: handler, User, []Course
	// - Return: okay / error
	amaterasu.With(ParseHandler(false)).With(ParseUser).Post("/create_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Start task
	// - Form: ID
	// - Return: okay / error
	amaterasu.Post("/start_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Stop task
	// ...
	amaterasu.Post("/stop_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Restart task
	// ...
	amaterasu.Post("/restart_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Delete task
	// ...
	amaterasu.Post("/delete_task", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Start all tasks
	// - Return: okay / error
	amaterasu.Post("/start_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Stop all tasks
	// ...
	amaterasu.Post("/stop_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Restart all tasks
	// ...
	amaterasu.Post("/restart_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Restart all tasks
	// ...
	amaterasu.Post("/delete_all", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	http.Handle("/", amaterasu.Handler())
}
