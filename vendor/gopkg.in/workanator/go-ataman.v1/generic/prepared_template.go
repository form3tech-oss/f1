package generic

import "fmt"

type preparedTemplate struct {
	tpl string
}

func (tpl preparedTemplate) String() string {
	return tpl.tpl
}

func (tpl preparedTemplate) Format(args ...interface{}) string {
	return fmt.Sprintf(tpl.tpl, args...)
}
