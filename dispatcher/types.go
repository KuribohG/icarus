package dispatcher

type SubtaskType int

const (
	SubtaskLogin SubtaskType = iota
	SubtaskList
	SubtaskElect
)

type Subtask struct {
	Handler string      `json:"handler"`
	Type    SubtaskType `json:"type"`
	Data    []string    `json:"data"`

	// Leave it empty and pass Subtask to dispatcher, and it will generate a random ID.
	ID int64 `json:"id"`
}

type TaskRequest struct {
	// Handlers that accepts.
	// If multiple handlers are indicated, separate them with comma (",").
	Accepts string `json:"accepts"`

	// Base64 encoded binary cipher.
	Cipher string `json:"cipher"`
}

type TaskResponse struct {
	// Is there any new task?
	OK bool `json:"ok"`

	// Encrypted Subtask, encoded in Base64.
	Content string `json:"content"`
}

type SubtaskResult struct {
	Data  []string `json:"data"`
	Error error    `json:"-"` // This member will be set by dispatcher if things go wrong.
}

type WorkResponse struct {
	// Encrypted SubtaskResult, encoded in Base64.
	Content string `json:"content"`

	// ID of the task.
	TaskID int64 `json:"task_id"`
}
