package basic

import (
	"gopkg.in/workanator/go-ataman.v1/ansi"
	"gopkg.in/workanator/go-ataman.v1/decorate"
)

var (
	tagOpen            = decorate.NewMarker("<")
	tagClose           = decorate.NewMarker(">")
	attributeDelimiter = decorate.NewMarker(",")
	modDelimiter       = decorate.NewMarker("+")
)

// Style returns the basic decoration style.
func Style() decorate.Style {
	return decorate.Style{
		TagOpen:            tagOpen,
		TagClose:           tagClose,
		AttributeDelimiter: attributeDelimiter,
		ModDelimiter:       modDelimiter,
		Attributes:         ansi.DefaultDict,
	}
}
