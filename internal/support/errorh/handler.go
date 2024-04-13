package errorh

import (
	"io"

	"github.com/sirupsen/logrus"
)

func Log(err error, message string) {
	if err != nil {
		logrus.WithError(err).Error(message)
	}
}

func SafeClose(closer io.Closer) {
	if closer == nil {
		return
	}

	Log(closer.Close(), "closed with error")
}
