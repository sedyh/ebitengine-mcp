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
	Error(err error)
}

type Failer struct {
	Err error
}

func (e *Failer) Error(err error) {
	e.Err = err
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
	Images []string
}

func (r *RecordResponse) Type() string {
	return "record-response"
}
