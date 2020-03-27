package ecslogs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"syscall"
	"time"
	"strconv"
)

type EventError struct {
	Type          string      `json:"type,omitempty"`
	Error         string      `json:"error,omitempty"`
	Errno         int         `json:"errno,omitempty"`
	Stack         interface{} `json:"stack,omitempty"`
	OriginalError error       `json:"origError,omitempty"`
}

func MakeEventError(err error) EventError {
	e := EventError{
		Type:          reflect.TypeOf(err).String(),
		Error:         err.Error(),
		OriginalError: err,
	}

	if errno, ok := err.(syscall.Errno); ok {
		e.Errno = int(errno)
	}

	return e
}

type EventInfo struct {
	Host   string       `json:"host,omitempty"`
	Source string       `json:"source,omitempty"`
	ID     string       `json:"id,omitempty"`
	PID    int          `json:"pid,omitempty"`
	UID    int          `json:"uid,omitempty"`
	GID    int          `json:"gid,omitempty"`
	Errors []EventError `json:"errors,omitempty"`
}

func (e EventInfo) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e EventInfo) String() string {
	return string(e.Bytes())
}

type EventData map[string]interface{}

func (e EventData) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e EventData) String() string {
	return string(e.Bytes())
}

type Event struct {
	Level   Level     `json:"level"`
	Time    time.Time `json:"time"`
	Info    EventInfo `json:"info"`
	Data    EventData `json:"data"`
	Message json.RawMessage `json:"message"`
	IsMessageString bool `json:"is_message_string"`
	WasMessagequoted bool `json:"was_message_quoted"`
}

// func Eprintf(level Level, format string, args ...interface{}) Event {
// 	return MakeEvent(level, sprintf(format, args...), args...)
// }

// func Eprint(level Level, args ...interface{}) Event {
// 	return MakeEvent(level, sprint(args...), args...)
// }

func stringToRawMessage(str string) (json.RawMessage, bool) {
    var js json.RawMessage
	err := json.Unmarshal([]byte(str), &js)
	return js, (err == nil)
}

func MakeEvent(level Level, message string, values ...interface{}) Event {
	var rawJsonMessage json.RawMessage
	var is_message_string bool
	var was_message_quoted bool
	var errors []EventError

	for _, val := range values {
		switch v := val.(type) {
		case error:
			errors = append(errors, MakeEventError(v))
		}
	}

	if raw, ok := stringToRawMessage(message); ok {
		if unquoted, err :=  strconv.Unquote(message); err == nil {
			if raw1, ok1 := stringToRawMessage(unquoted); ok1 {
				rawJsonMessage = raw1
				is_message_string = false
				was_message_quoted = true
			} else {
				rawJsonMessage = raw
				is_message_string = false
				was_message_quoted = false
			}
		} else {
			rawJsonMessage = raw
			is_message_string = false
			was_message_quoted = false
		}
	} else {
		string_raw, _ := json.Marshal(message)
		rawJsonMessage = json.RawMessage(string(string_raw))
		is_message_string = true
		was_message_quoted = false
	}


	return Event{
		Info:    EventInfo{Errors: errors},
		Data:    EventData{},
		Level:   level,
		Message: rawJsonMessage,
		IsMessageString: is_message_string,
		WasMessagequoted: was_message_quoted,
	}
}

func (e Event) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e Event) String() string {
	return string(e.Bytes())
}

func copyEventData(data ...EventData) EventData {
	copy := EventData{}

	for _, d := range data {
		for k, v := range d {
			copy[k] = v
		}
	}

	return copy
}

func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func sprint(args ...interface{}) string {
	s := fmt.Sprintln(args...)
	return s[:len(s)-1]
}
