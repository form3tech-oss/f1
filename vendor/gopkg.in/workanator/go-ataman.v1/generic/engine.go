package generic

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/workanator/go-ataman.v1/ansi"
)

// Reset graphic mode sequence
var ansiResetSequence = fmt.Sprintf("%s%d%s", ansi.SequenceStart, ansi.Reset, ansi.SequenceEnd)

// Render renders the template given.
func (rndr *Renderer) renderTemplate(tpl *string, buf stringWriter) error {
	var (
		openSeq     = rndr.style.TagOpen.String()
		closeSeq    = rndr.style.TagClose.String()
		doubleClose = strings.Repeat(closeSeq, 2)
		writtenANSI = false
	)

	pos := 0
	for pos < len(*tpl) {
		// Search for tag open marker
		idx := strings.Index((*tpl)[pos:], openSeq)

		// Write the chunk from the position till the open marker or the end of line
		var endPos int
		if idx > 0 {
			endPos = pos + idx
		} else if idx == -1 {
			endPos = len(*tpl)
		}

		if endPos > 0 {
			// Convert all doubled tag close sequences
			for {
				doubleCloseIdx := strings.Index((*tpl)[pos:endPos], doubleClose)
				if doubleCloseIdx >= 0 {
					n, _ := buf.WriteString((*tpl)[pos : pos+doubleCloseIdx])
					pos += n

					n, _ = buf.WriteString(closeSeq)
					pos += n * 2
				} else {
					break
				}
			}

			if pos < endPos {
				n, _ := buf.WriteString((*tpl)[pos:endPos])
				pos += n
			}
		}

		// Convert tag if any into ANSI sequence
		if idx >= 0 {
			openIdx := pos

			// Check if open tag sequence is doubled
			secondOpenIdx := strings.Index((*tpl)[openIdx+1:], openSeq)
			if secondOpenIdx == 0 {
				buf.WriteString(openSeq)
				pos += len(openSeq) * 2
			} else {
				// Find the closing tag position
				closingIdx := strings.Index((*tpl)[openIdx:], closeSeq)
				if closingIdx >= 0 {
					pos += closingIdx + 1
				} else {
					return errors.New("attribute tag misses closing sequence")
				}

				// Get tag content and split to attribute list
				content := (*tpl)[openIdx+1 : openIdx+closingIdx]
				attributes := strings.Split(content, rndr.style.AttributeDelimiter.String())

				// Build ANSI sequence from attributes
				sequnce, err := rndr.ansiSequence(attributes)
				if err != nil {
					return err
				}

				buf.WriteANSISequence(sequnce)
				writtenANSI = true
			}
		}
	}

	if writtenANSI {
		buf.WriteANSISequence(ansiResetSequence)
	}

	return nil
}

// ansiSequence produces ANSI sequence based on attribute list.
func (rndr *Renderer) ansiSequence(attrs []string) (string, error) {
	for i, attr := range attrs {
		code := rndr.ansiCode(attr)
		if !code.IsValid() {
			return "", fmt.Errorf("invalid attribute %s", attr)
		}

		attrs[i] = fmt.Sprintf("%d", code)
	}

	return ansi.SequenceStart + strings.Join(attrs, ansi.SequenceDelimiter) + ansi.SequenceEnd, nil
}

// ansiCode returns the ANSI numeric code of the attribute.
func (rndr *Renderer) ansiCode(attr string) ansi.Attribute {
	var code ansi.Attribute

	mods := strings.Split(attr, rndr.style.ModDelimiter.String())
	for _, mod := range mods {
		if a, ok := rndr.style.Attributes[mod]; ok {
			code += a
		} else {
			return ansi.InvalidAttribute
		}
	}

	return code
}
