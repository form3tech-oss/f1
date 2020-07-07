package generic

import (
	"fmt"
	"sync"

	"gopkg.in/workanator/go-ataman.v1/decorate"
	"gopkg.in/workanator/go-ataman.v1/prepared"
)

// Renderer implements generic configurable template renderer. Underlying pool
// is used for pooling string buffers and make the renderer thread safe.
type Renderer struct {
	style decorate.Style
	pool  sync.Pool
	mutex sync.Mutex
}

// NewRenderer constructs the instance of generic renderer with the decoration
// style given.
func NewRenderer(style decorate.Style) *Renderer {
	return &Renderer{
		style: style,
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytesBuffer)
			},
		},
		mutex: sync.Mutex{},
	}
}

// Validate validates the template.
func (rndr *Renderer) Validate(tpl string) error {
	var buf mockBuffer
	return rndr.renderTemplate(&tpl, &buf)
}

// Render renders the template given.
func (rndr *Renderer) Render(tpl string) (string, error) {
	buf := rndr.getBuffer()
	defer rndr.putBuffer(buf)

	err := rndr.renderTemplate(&tpl, buf)

	return buf.String(), err
}

// MustRender renders the template and panics in case of error.
func (rndr *Renderer) MustRender(tpl string) string {
	buf := rndr.getBuffer()
	defer rndr.putBuffer(buf)

	if err := rndr.renderTemplate(&tpl, buf); err != nil {
		panic(err)
	}

	return buf.String()
}

// Renderf formats and renders the template given.
func (rndr *Renderer) Renderf(tpl string, args ...interface{}) (string, error) {
	return rndr.Render(fmt.Sprintf(tpl, args...))
}

// MustRenderf formats and renders the template and panics in case of error.
func (rndr *Renderer) MustRenderf(tpl string, args ...interface{}) string {
	result, err := rndr.Renderf(tpl, args...)
	if err != nil {
		panic(err)
	}

	return result
}

// Len returns the length of the text the user see in terminal.
func (rndr *Renderer) Len(tpl string) int {
	var buf mockBuffer
	var err = rndr.renderTemplate(&tpl, &buf)
	if err != nil {
		return len(tpl)
	}

	return buf.Len()
}

// Lenf calculates and return the length of the formatted template.
func (rndr *Renderer) Lenf(tpl string, args ...interface{}) int {
	return rndr.Len(fmt.Sprintf(tpl, args...))
}

// Prepare prerenders the template given.
func (rndr *Renderer) Prepare(tpl string) (prepared.Template, error) {
	buf := rndr.getBuffer()
	defer rndr.putBuffer(buf)

	err := rndr.renderTemplate(&tpl, buf)
	if err != nil {
		return nil, err
	}

	return preparedTemplate{tpl: buf.String()}, nil
}

// MustPrepare prerenders the template and panics in case of parsing error.
func (rndr *Renderer) MustPrepare(tpl string) (pt prepared.Template) {
	buf := rndr.getBuffer()
	defer rndr.putBuffer(buf)

	err := rndr.renderTemplate(&tpl, buf)
	if err != nil {
		panic(err)
	}

	return preparedTemplate{tpl: buf.String()}
}

func (rndr *Renderer) getBuffer() *bytesBuffer {
	rndr.mutex.Lock()
	defer rndr.mutex.Unlock()

	return rndr.pool.Get().(*bytesBuffer)
}

func (rndr *Renderer) putBuffer(buf *bytesBuffer) {
	rndr.mutex.Lock()
	defer rndr.mutex.Unlock()

	buf.Buffer.Reset()
	rndr.pool.Put(buf)
}
