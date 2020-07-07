package ansi

const (
	// SequenceStart is ANSI graphic mode sequence start
	SequenceStart = "\033["
	// SequenceEnd is ANSI graphic mode sequence end
	SequenceEnd = "m"
	// SequenceDelimiter is the delimiter of ANSI graphic modes
	SequenceDelimiter = ";"

	// Reset makes all attributes off
	Reset Attribute = 0

	// Bold makes font bold (or brighter)
	Bold Attribute = 1
	// Underscore (or underline) makes font underline.
	Underscore Attribute = 4
	// Blink makes font blink
	Blink Attribute = 5
	// Reverse reverses tex and background colors
	Reverse Attribute = 7
	// Concealed makes text comcealed
	Concealed Attribute = 8

	// Black makes text color black
	Black Attribute = 30
	// Red makes text color red
	Red Attribute = 31
	// Green makes text color green
	Green Attribute = 32
	// Yellow makes text color yellow
	Yellow Attribute = 33
	// Blue makes text color blue
	Blue Attribute = 34
	// Magenta makes text color magenta
	Magenta Attribute = 35
	// Cyan makes text color cyan
	Cyan Attribute = 36
	// White makes text color white
	White Attribute = 37
	// Default resets text color to the default color
	Default Attribute = 39

	// Background when added to text color makes the result mode which has effect
	// on the background, e.g. `Red + Background` makes red background ANSI mode.
	Background Attribute = 10

	// Intensive when added to text color makes the result mode to have high
	// intensive color, e.g. `Blue + Intensive` makes high intensive blue text
	// color. Can be used with Background like `Cyan + Background + Intensive`.
	Intensive Attribute = 60

	// Light is the synonym to Intensive
	Light Attribute = 60
)
