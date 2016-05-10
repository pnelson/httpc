package httpc

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/pnelson/tmpl"
)

// Viewable represents a view. To provide an expressive API, this
// type is an alias for interface{} that is named for documentation.
type Viewable interface{}

// Render writes the view in the requested format, if available.
func (ctx *Context) Render(view Viewable, code int) error {
	accept := ctx.Request.Header.Get("Accept")
	if accept == "" {
		return ctx.RenderJSON(view, code)
	}
	for _, h := range strings.Split(accept, ",") {
		media, _, err := mime.ParseMediaType(h)
		if err != nil {
			return err
		}
		switch media {
		case "text/html", "text/*":
			if renderer == nil {
				continue
			}
			v, ok := view.(tmpl.Viewable)
			if ok {
				return ctx.RenderHTML(v, code)
			}
		case "application/json", "application/*", "*/*":
			return ctx.RenderJSON(view, code)
		case "text/plain":
			return ctx.RenderPlain(view, code)
		}
	}
	return ctx.Abort(http.StatusNotAcceptable)
}

// Renderer represents the ability to render a tmpl.Viewable.
type Renderer interface {
	Render(view tmpl.Viewable) ([]byte, error)
}

// renderer is used to render HTML templates.
var renderer Renderer

// SetRenderer sets the Renderer used to render HTML templates.
// This function is not thread safe and is intended to be called
// once during application initialization.
func SetRenderer(r Renderer) {
	mu.Lock()
	renderer = r
	mu.Unlock()
}

// RenderHTML writes the view as templated HTML.
func (ctx *Context) RenderHTML(view tmpl.Viewable, code int) error {
	b, err := renderer.Render(view)
	if err != nil {
		return err
	}
	ctx.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.WriteHeader(code)
	if view == nil {
		return nil
	}
	_, err = ctx.Write(b)
	return err
}

// RenderJSON writes the view as marshalled JSON.
func (ctx *Context) RenderJSON(view Viewable, code int) error {
	b, err := json.Marshal(view)
	if err != nil {
		return err
	}
	ctx.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.WriteHeader(code)
	if view == nil {
		return nil
	}
	_, err = ctx.Write(b)
	return err
}

// RenderPlain writes the view as a string.
func (ctx *Context) RenderPlain(view Viewable, code int) error {
	s, ok := view.(string)
	if !ok {
		return fmt.Errorf("httpc: view for RenderPlain must be a string")
	}
	ctx.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Header().Set("X-Content-Type-Options", "nosniff")
	ctx.WriteHeader(code)
	_, err := fmt.Fprintln(ctx, s)
	return err
}
