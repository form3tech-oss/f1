package ataman

import (
	"gopkg.in/workanator/go-ataman.v1/decorate"
	"gopkg.in/workanator/go-ataman.v1/generic"
)

// NewRenderer creates the generic configurable renderer instance with
// the decoration style given.
func NewRenderer(style decorate.Style) Renderer {
	return generic.NewRenderer(style)
}
