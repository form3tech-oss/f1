package trace

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/form3tech-oss/f1/v2/internal/termcolor"
)

type ConsoleTracer struct {
	writer io.Writer
}

func NewConsoleTracer(output io.Writer) *ConsoleTracer {
	return &ConsoleTracer{writer: output}
}

var _ Tracer = &ConsoleTracer{}

func colorString(s string, c string) string {
	return c + s + termcolor.Reset
}

func (t *ConsoleTracer) ReceivedFromChannel(name string) {
	t.event("Received from Channel '%s'", name)
}

func (t *ConsoleTracer) SentToChannel(name string) {
	t.event("Sent to Channel '%s'", name)
}

func (t *ConsoleTracer) SendingToChannel(name string) {
	t.event("Sending to Channel '%s'", name)
}

func (t *ConsoleTracer) Event(message string, args ...any) {
	t.event(message, args...)
}

func (t *ConsoleTracer) event(message string, args ...any) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	keywords := []string{"channel", "Channel"}

	fMessage := colorString(fmt.Sprintf(message, args...), termcolor.Yellow)

	for _, s := range keywords {
		fMessage = strings.Replace(fMessage, s, termcolor.Red+s+termcolor.Yellow, 1)
	}

	fileName := frame.File + strconv.Itoa(frame.Line)
	fileName = colorString(fileName, termcolor.Blue)

	now := time.Now()
	formattedTime := fmt.Sprintf("%02d:%02d:%02d %02dms",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/int(time.Millisecond))

	timePart := colorString(fmt.Sprintf("[TRACE - %s]: ", formattedTime), termcolor.White)
	messagePart := colorString(fMessage+" -> ", termcolor.White)
	filePart := colorString(fileName, termcolor.Blue)

	fmt.Fprintln(t.writer, timePart+messagePart+filePart)
}
