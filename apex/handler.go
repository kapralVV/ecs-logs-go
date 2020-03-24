package apex_ecslogs

import (
	"encoding/json"
	"strconv"
	"io"
	"fmt"

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

		if pc, ok := ecslogs.GuessCaller(c.Depth, 10, "github.com/segmentio/ecs-logs", "github.com/apex/log"); ok {
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

func IsJSON(str string) bool {
    var js json.RawMessage
    return json.Unmarshal([]byte(str), &js) == nil
}

func makeEvent(entry *apex.Entry, source string) ecslogs.Event {
	var message json.RawMessage
	if IsJSON(entry.Message) {
		if unquotEdStr, err := strconv.Unquote(entry.Message); err == nil {
			if IsJSON(unquotEdStr) {
				message = json.RawMessage(unquotEdStr)
			} else {
				message = json.RawMessage(fmt.Sprintf(`{"string": %s}`, strconv.Quote(entry.Message)))
			}
		} else {
			message = json.RawMessage(fmt.Sprintf(`{"string": %s}`, strconv.Quote(entry.Message)))
		}
	} else {
		message = json.RawMessage(fmt.Sprintf(`{"string": %s}`, strconv.Quote(entry.Message)))
	}
	return ecslogs.Event{
		Level:   makeLevel(entry.Level),
		Info:    makeEventInfo(entry, source),
		Data:    makeEventData(entry),
		Time:    entry.Timestamp,
		Message: message,
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
