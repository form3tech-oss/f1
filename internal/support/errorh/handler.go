package errorh

import (
	"io"

	log "github.com/sirupsen/logrus"
)

func Log(err error, message string) {
	if err != nil {
		log.WithError(err).Error(message)
	}
}

func SafeClose(closer io.Closer) {
	if closer == nil {
		return
	}

	Log(closer.Close(), "closed with error")
}
