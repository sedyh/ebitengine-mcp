package event

import (
	"fmt"
	"strings"
	"time"
)

func Mark(id, cat string) (marked string) {
	return fmt.Sprintf("%s/%s", id, cat)
}

func Unmark(marked string) (id, cat string) {
	parts := strings.Split(marked, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

type Event interface {
	Type() string
	SetError(err error)
}

type Failer struct {
	err error
}

func (e *Failer) SetError(err error) {
	e.err = err
}

func (e *Failer) Error() error {
	return e.err
}

type Logger struct {
	logs []string
}

func (e *Logger) Logs() []string {
	return e.logs
}

func (e *Logger) SetLogs(logs []string) {
	e.logs = logs
}

type RecordRequest struct {
	Failer
	Target string
	Frames int
	Delay  time.Duration
}

func (r *RecordRequest) Type() string {
	return "record-request"
}

type RecordResponse struct {
	Failer
	Logger
	Images []string
}

func (r *RecordResponse) Type() string {
	return "record-response"
}
