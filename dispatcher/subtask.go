package dispatcher

type Subtask struct {
	Handler string `json:"handler"`
	Type    string `json:"type"`
	Data    string `json:"data"`
}
