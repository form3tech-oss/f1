package logutils

import (
	"github.com/form3tech-oss/f1/v3/internal/envsettings"
	"github.com/form3tech-oss/f1/v3/internal/log"
)

func NewLogConfigFromSettings(settings envsettings.Settings) *log.Config {
	return log.NewConfig().
		WithLevel(settings.Log.SlogLevel()).
		WithJSONFormat(settings.Log.IsFormatJSON())
}
