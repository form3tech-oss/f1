package trace

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

var reset = "\033[0m"
var gray = "\033[37m"
var yellow = "\033[33m"
var blue = "\033[34m"
var red = "\033[31m"

func colorString(s string, c string) string {
	return c + s + reset
}

func ReceivedFromChannel(name string) {
	event("Received from Channel '%s'", name)
}

func SentToChannel(name string) {
	event("Sent to Channel '%s'", name)
}

func SendingToChannel(name string) {
	event("Sending to Channel '%s'", name)
}

func Event(message string, args ...interface{}) {
	event(message, args...)
}

func event(message string, args ...interface{}) {
	if os.Getenv("TRACE") != "true" {
		return
	}

	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	keywords := []string{"channel", "Channel"}

	fMessage := colorString(fmt.Sprintf(message, args...), yellow)

	for _, s := range keywords {
		fMessage = strings.Replace(fMessage, s, red+s+yellow, 1)
	}

	fileName := strings.Replace(frame.File, os.Getenv("GOPATH")+"/src/github.com/form3tech/k6-tests", ".", 1)
	fileName = fmt.Sprintf("%s:%d", fileName, frame.Line)
	fileName = colorString(fileName, blue)

	t := time.Now()
	formattedTime := fmt.Sprintf("%02d:%02d:%02d %02dms",
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/int(time.Millisecond))

	timePart := colorString(fmt.Sprintf("[TRACE - %s]: ", formattedTime), gray)
	messagePart := colorString(fmt.Sprintf("%s -> ", fMessage), gray)
	filePart := colorString(fileName, blue)

	fmt.Println(timePart + messagePart + filePart)
}
