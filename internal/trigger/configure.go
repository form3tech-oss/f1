package trigger

import (
	"github.com/form3tech-oss/f1/v2/internal/trigger/api"
	"github.com/form3tech-oss/f1/v2/internal/trigger/constant"
	"github.com/form3tech-oss/f1/v2/internal/trigger/file"
	"github.com/form3tech-oss/f1/v2/internal/trigger/gaussian"
	"github.com/form3tech-oss/f1/v2/internal/trigger/ramp"
	"github.com/form3tech-oss/f1/v2/internal/trigger/staged"
	"github.com/form3tech-oss/f1/v2/internal/trigger/users"
	"github.com/form3tech-oss/f1/v2/internal/ui"
)

func GetBuilders(output *ui.Output) []api.Builder {
	return []api.Builder{
		constant.Rate(),
		staged.Rate(),
		gaussian.Rate(output),
		users.Rate(),
		ramp.Rate(),
		file.Rate(output),
	}
}
