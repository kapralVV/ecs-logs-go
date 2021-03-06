package apex_ecslogs

import (
	"bytes"
	"io"
	"strings"
	"testing"

	apex "github.com/apex/log"
	"github.com/segmentio/ecs-logs-go"
)

func TestHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	log := &apex.Logger{
		Handler: NewHandlerWith(Config{
			Output:   buf,
			FuncInfo: testFuncInfo,
		}),
		Level: apex.DebugLevel,
	}

	log.
		WithField("error", io.EOF).
		WithField("hello", "world").
		Errorf("an error was raised (%s)", io.EOF)

	s := strings.TrimSpace(buf.String())

	// I wish we could make better testing here but the apex
	// API doesn't let us mock the timestamp so we can't really
	// predict what "time" is gonna be.
	if !strings.HasPrefix(s, `{"level":"ERROR","time":"`) || !strings.HasSuffix(s, `"errors":[{"type":"*errors.errorString","error":"EOF","origError":{}}]},"data":{"hello":"world"},"message":"an error was raised (EOF)"}`) {
		t.Error("apex handler failed:", s)
	}
}

func testFuncInfo(pc uintptr) (info ecslogs.FuncInfo, ok bool) {
	if info, ok = ecslogs.GetFuncInfo(pc); !ok {
		return
	}
	info.Line = 42
	return
}
