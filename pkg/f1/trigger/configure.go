package trigger

import (
	"github.com/form3tech-oss/f1/pkg/f1/trigger/api"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/constant"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/file"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/gaussian"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/ramp"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/staged"
	"github.com/form3tech-oss/f1/pkg/f1/trigger/users"
)

func GetBuilders() []api.Builder {
	return []api.Builder{
		constant.ConstantRate(),
		staged.StagedRate(),
		gaussian.GaussianRate(),
		users.UsersRate(),
		ramp.RampRate(),
		file.FileRate(),
	}
}
