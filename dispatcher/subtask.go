package dispatcher

type SubtaskType int

const (
	SubtaskLogin SubtaskType = iota
	SubtaskList
	SubtaskElect
)

type Subtask struct {
	Handler string   `json:"handler"`
	Type    string   `json:"type"`
	Data    []string `json:"data"`
}
