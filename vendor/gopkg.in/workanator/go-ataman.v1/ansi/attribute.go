package ansi

// Attribute is the numerc code used in ANSI sequence.
type Attribute int

// InvalidAttribute identifies improperly defined ANSI codes.
const InvalidAttribute Attribute = -1

// IsValid tests if the attribute is valid ANSI code.
func (attr Attribute) IsValid() bool {
	return attr != InvalidAttribute
}
