package apex_ecslogs

import (
	"encoding/json"
	"strconv"
	"io"

	apex "github.com/apex/log"
	"github.com/kapralVV/ecs-logs-go"
)

type Config struct {
	Output   io.Writer
	Depth    int
	FuncInfo func(uintptr) (ecslogs.FuncInfo, bool)
}

func NewHandler(w io.Writer) apex.Handler {
	return NewHandlerWith(Config{Output: w})
}

func NewHandlerWith(c Config) apex.Handler {
	logger := ecslogs.NewLogger(c.Output)

	if c.FuncInfo == nil {
		return apex.HandlerFunc(func(entry *apex.Entry) error {
			return logger.Log(MakeEvent(entry))
		})
	}

	return apex.HandlerFunc(func(entry *apex.Entry) error {
		var source string

		if pc, ok := ecslogs.GuessCaller(c.Depth, 10, "github.com/kapralVV/ecs-logs", "github.com/apex/log"); ok {
			if info, ok := c.FuncInfo(pc); ok {
				source = info.String()
			}
		}

		return logger.Log(makeEvent(entry, source))
	})
}

func MakeEvent(entry *apex.Entry) ecslogs.Event {
	return makeEvent(entry, "")
}

func stringToRawMessage(str string) (json.RawMessage, bool) {
    var js json.RawMessage
	err := json.Unmarshal([]byte(str), &js)
	return js, (err == nil)
}

func makeEvent(entry *apex.Entry, source string) ecslogs.Event {
	var message json.RawMessage
	var isString bool
	var isQuoted bool

	raw, ok := stringToRawMessage(entry.Message)
	if ok {
		if unquoted, err :=  strconv.Unquote(entry.Message); err == nil {
			if raw1, ok1 := stringToRawMessage(unquoted); ok1 {
				message = raw1
				isString = false
				isQuoted = true
			} else {
				message = raw
				isQuoted = false
				isString = false
			}
		} else {
			message = raw
			isQuoted = false
			isString = false
		}
	} else {
		string_raw, _ := json.Marshal(entry.Message)
		message = json.RawMessage(string(string_raw))
		isString = true
		isQuoted = false
	}

	return ecslogs.Event{
		Level:   makeLevel(entry.Level),
		Info:    makeEventInfo(entry, source),
		Data:    makeEventData(entry),
		Time:    entry.Timestamp,
		Message: message,
		IsMessageString: isString,
		WasMessagequoted: isQuoted,
	}
}

func makeEventInfo(entry *apex.Entry, source string) ecslogs.EventInfo {
	return ecslogs.EventInfo{
		Source: source,
		Errors: makeErrors(entry.Fields),
	}
}

func makeEventData(entry *apex.Entry) ecslogs.EventData {
	data := make(ecslogs.EventData, len(entry.Fields))

	for k, v := range entry.Fields {
		data[k] = v
	}

	return data
}

func makeLevel(level apex.Level) ecslogs.Level {
	switch level {
	case apex.DebugLevel:
		return ecslogs.DEBUG

	case apex.InfoLevel:
		return ecslogs.INFO

	case apex.WarnLevel:
		return ecslogs.WARN

	case apex.ErrorLevel:
		return ecslogs.ERROR

	case apex.FatalLevel:
		return ecslogs.CRIT

	default:
		return ecslogs.NONE
	}
}

func makeErrors(fields apex.Fields) (errors []ecslogs.EventError) {
	for k, v := range fields {
		if err, ok := v.(error); ok {
			errors = append(errors, ecslogs.MakeEventError(err))
			delete(fields, k)
		}
	}
	return
}
