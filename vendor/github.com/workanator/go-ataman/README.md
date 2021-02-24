![Gopher Ataman](https://github.com/workanator/repo-faces/blob/main/go-ataman/main.png)

# go-ataman
**Another Text Attribute Manipulator**

[![GoDoc](https://godoc.org/gopkg.in/workanator/go-ataman.v1?status.svg)](https://godoc.org/gopkg.in/workanator/go-ataman.v1)
[![Build Status](https://travis-ci.org/workanator/go-ataman.svg?branch=master)](https://travis-ci.org/workanator/go-ataman)
[![Coverage Status](https://coveralls.io/repos/github/workanator/go-ataman/badge.svg?branch=master)](https://coveralls.io/github/workanator/go-ataman?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/workanator/go-ataman)](https://goreportcard.com/report/github.com/workanator/go-ataman)
[![License](https://img.shields.io/dub/l/vibe-d.svg)](https://github.com/workanator/go-ataman/blob/master/LICENSE)

The goal of the project is to help render colored text in terminal applications
with text attribute manipulation contained in text template.
The idea is similar to HTML, e.g. `<span style="color:green">Hello!</span>`

**ataman** is not a full-featured template processor. It aims only on simple
terminal text attribute, graphic mode in ANSI terms, manipulations using
ANSI sequences.

## Features

- Parsing and rendering template with single iteration.
- Minimal memory allocations with pooling.
- Thread-safe parsing and rendering.
- Fully customizable decoration of tags.

## Installation

To install the package use `go get gopkg.in/workanator/go-ataman.v1`

## Quick Example

The next example shows basic usage of the package. The renderer here uses
basic decoration style.

```go
rndr := ataman.NewRenderer(ataman.BasicStyle())
tpl := "<light+green>%s<->, <bg+light+yellow,black,bold> <<%s>> <-><red>!"
fmt.Println(rndr.MustRenderf(tpl, "Hello", "Terminal World"))

```

This example produces colored text like this.

![Example Output](https://github.com/workanator/repo-faces/blob/main/go-ataman/example1.png)

Renderer is able to pre-render templates so in further output operations they
can be reused without parsing and rendering. In this example the renderer uses
curly brackets decoration style.

```go
rndr := ataman.NewRenderer(ataman.CurlyStyle())
prep := rndr.MustPrepare("{light_green}%s{-}, {bg_light_yellow+black+bold} <%s> {-}{red}!")
fmt.Println(rndr.MustRenderf(tpl, "Hi", "Everybody"))
```

The pre-rendered template implements `fmt.Stringer` interface so it can be used
in output operations without additional code, e.g.

```go
prep := rndr.MustPrepare("{red}Red {green}Green {blue}Blue{-}")
fmt.Println(prep)
```

## Writing Templates

Templates follow the simple rules.

- Tag can contain one or more attributes, e.g. `{white,bold}`, `{red}`.
- If template should render open or close tag sequence as regular text then
  the sequence should be doubled. For example, if tag is enclosed in `{` and `}`
  then in template it should be `This rendered as {{regular text}}`.
- Renderer adds reset graphic mode ANSI sequence to the each template if it
  contains any other ANSI sequences. So visually those templates are equivalent
  `{bold}Bold{-}` and `{bold}Bold`.

Decoration styles use the follows dictionary.

* `-` or `reset` stand for reset graphic mode.
* `b` or `bold` make font bold.
* `u` or `underscore` or `underline` make font underline.
* `blink` makes font blink.
* `reverse` swaps text and background colors.
* `conceal` makes font concealed (whatever that means).
* `black` color.
* `red` color.
* `green` color.
* `yellow` color.
* `blue` color.
* `magenta` color.
* `cyan` color.
* `white` color.
* `default` reverts to the default color.
* `bg` or `background` should be used in conjunction with color to set
  background color.
* `intensive` or `light` should be used in conjunction with color to make
  the color intensive. Could be used with `background` as well.

Some template examples with curly decoration style.

- `{light_green}` - makes text light (intensive) green.
- `{bg_yellow}` - makes background yellow.
- `{bold}` - makes font bold.
- `{red,bg_blue}` - makes text red on blue background.
- `{u,black,bg_intensive_white}` - makes text black with underline font on
  intensive white background.
- `{-}` - reset the current graphic mode.

## Customizing Renderer

The package allows to customize tag decorations what can be achieved through
`decorate.Style` struct. The struct should be initialized with preferred
values. For example with the code below we can define a decoration style
like `[[bold,yellow]]Warning![[-]] [[intensive_white]]This package is awesome![[-]] :)`.

```go
style := decorate.Style{
  TagOpen:            decorate.NewMarker("[["),
  TagClose:           decorate.NewMarker("]]"),
  AttributeDelimiter: decorate.NewMarker(","),
  ModDelimiter:       decorate.NewMarker("-"),
  Attributes:         ansi.DefaultDict,
}

rndr := ataman.NewRenderer(style)
```

The rules of decoration styles are the follows.

- `TagOpen` is the sequence of runes which open tag.
- `TagClose` is the sequence of runes which close tag.
- `AttributeDelimiter` is the sequence of runes which delimits attributes
  inside tag.
- `ModDelimiter` is the sequence of runes which delimits modifiers
  in attribute.
- `Attributes` is the map of attributes, where key is the name of ANSI code
  user uses in templates and value is the ANSI code used in ANSI sequence.

It's recommended to use attribute codes defined in `ansi` package with the
default Renderer provided by **ataman**.
