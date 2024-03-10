package trace

import (
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	termReset  = "\033[0m"
	termGray   = "\033[37m"
	termYellow = "\033[33m"
	termBlue   = "\033[34m"
	termRed    = "\033[31m"
)

type ConsoleTracer struct {
	writer io.Writer
}

func NewConsoleTracer(output io.Writer) *ConsoleTracer {
	return &ConsoleTracer{writer: output}
}

var _ Tracer = &ConsoleTracer{}

func colorString(s string, c string) string {
	return c + s + termReset
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

	fMessage := colorString(fmt.Sprintf(message, args...), termYellow)

	for _, s := range keywords {
		fMessage = strings.Replace(fMessage, s, termRed+s+termYellow, 1)
	}

	fileName := frame.File + strconv.Itoa(frame.Line)
	fileName = colorString(fileName, termBlue)

	now := time.Now()
	formattedTime := fmt.Sprintf("%02d:%02d:%02d %02dms",
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/int(time.Millisecond))

	timePart := colorString(fmt.Sprintf("[TRACE - %s]: ", formattedTime), termGray)
	messagePart := colorString(fMessage+" -> ", termGray)
	filePart := colorString(fileName, termBlue)

	fmt.Fprintln(t.writer, timePart+messagePart+filePart)
}
