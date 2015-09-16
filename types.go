package icarus

type LoginSession interface{}
type Course interface {
	Name() string
	Elect(LoginSession) (bool, error)
}
type User interface {
	Name() string
	Login() (LoginSession, error)
	ListCourse() ([]CourseData, error)
}

type UserData struct {
	UserID   string `json:"userid"`
	Password string `json:"password,omitempty"`
}
type CourseData struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Token string `json:"token,omitempty"`
}
type TaskData struct {
	ID      int          `json:"id"`
	Handle  string       `json:"handle"`
	User    UserData     `json:"user"`
	Courses []CourseData `json:"courses"`
	Stat    Stat         `json:"stat,omitempty"`
}
type Stat struct {
	Running   bool   `json:"running"`
	Succeeded int64  `json:"succeeded"`
	Failed    int64  `json:"failed"`
	LastError string `json:"last_error"`
	Elected   bool   `json:"elected"`
}
