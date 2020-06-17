package ansi

// DefaultDict is the default dictionary on ANSI attributes used in styles
// defined in the package.
var DefaultDict = map[string]Attribute{
	"-":          Reset,
	"reset":      Reset,
	"b":          Bold, // Font attributes
	"bold":       Bold,
	"u":          Underscore,
	"underscore": Underscore,
	"underline":  Underscore,
	"blink":      Blink,
	"reverse":    Reverse,
	"conceal":    Concealed,
	"black":      Black, // Text colors
	"red":        Red,
	"green":      Green,
	"yellow":     Yellow,
	"blue":       Blue,
	"magenta":    Magenta,
	"cyan":       Cyan,
	"white":      White,
	"default":    Default,
	"bg":         Background, // Background modificator
	"background": Background,
	"intensive":  Intensive, // High intensive modificator
	"light":      Light,
}
