/*
Package ataman is a colored terminal text rendering engine using ANSI sequences.
The project aims on simple text attribute manipulations with templates.

Here is the couple of examples to introduce the project.

  // Rendering with the renderer
  rndr := ataman.NewRenderer(ataman.CurlyStyle())
  tpl := "{red}Red {green}Green {blue}Blue{-}"
  fmt.Println(rndr.MustRender(tpl))

  // Rendering pre-rendered template.
  prep := rndr.MustPrepare("{light_green}%s{-}, {bg_light_yellow+blue+bold}%s{-}!")
  fmt.Println(prep.Format("Hello", "World"))

Customization of decoration styles can be done through `decorate.Style`, e.g.

  style := decorate.Style{
    TagOpen:            decorate.NewMarker("["),
    TagClose:           decorate.NewMarker("]"),
    AttributeDelimiter: decorate.NewMarker(","),
    ModDelimiter:       decorate.NewMarker("-"),
    Attributes:         ansi.DefaultDict,
  }

  rndr := ataman.NewRenderer(style)
  tpl := "[bold,yellow]Warning![-] [intensive_white]This package is awesome![-]"
  fmt.Println(rndr.MustRender(tpl))

Templates follow the simple rules.

  - Tag can contain one or more attributes, e.g. `{white,bold}`, `{red}`.
  - If template should render open or close tag sequence as regular text then
    the sequence should be doubled. For example, if tag is enclosed in `{` and `}`
    then in template it should be `This rendered as {{regular text}}`.
  - Renderer adds reset graphic mode ANSI sequence to the each template if it
    contains any other ANSI sequences. So visually those templates are equivalent
    `{bold}Bold{-}` and `{bold}Bold`.

Decoration styles use the follows dictionary.

  - `-` or `reset` stand for reset graphic mode.
  - `b` or `bold` make font bold.
  - `u` or `underscore` or `underline` make font underline.
  - `blink` makes font blink.
  - `reverse` swaps text and background colors.
  - `conceal` makes font concealed (whatever that means).
  - `black` color.
  - `red` color.
  - `green` color.
  - `yellow` color.
  - `blue` color.
  - `magenta` color.
  - `cyan` color.
  - `white` color.
  - `default` reverts to the default color.
  - `bg` or `background` should be used in conjunction with color to set
    background color.
  - `intensive` or `light` should be used in conjunction with color to make
    the color intensive. Could be used with `background` as well.

Some template examples with curly decoration style.

  - `{light_green}` - makes text light (intensive) green.
  - `{bg_yellow}` - makes background yellow.
  - `{bold}` - makes font bold.
  - `{red,bg_blue}` - makes text red on blue background.
  - `{u,black,bg_intensive_white}` - makes text black with underline font on
    intensive white background.
  - `{-}` - reset the current graphic mode.
*/
package ataman
