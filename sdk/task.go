package sdk

type Task struct {
	ID   string
	Body []byte
}

func (e *Task) Id() string { return e.ID }

func (e *Task) Event() string { return "task" }

func (e *Task) Data() string { return string(e.Body) }
