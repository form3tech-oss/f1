package errorh

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

func Log(err error, message string) {
	if err != nil {
		log.WithError(err).Error(message)
	}
}

func Print(err error, message string) {
	if err != nil {
		fmt.Printf("%s: %s", message, err)
	}
}

func SafeClose(closer io.Closer) {
	if closer == nil {
		return
	}

	Log(closer.Close(), "closed with error")
}
