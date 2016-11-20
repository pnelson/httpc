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
func Render(w http.ResponseWriter, req *http.Request, view Viewable, code int) error {
	accept := req.Header.Get("Accept")
	if accept == "" {
		return RenderJSON(w, view, code)
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
				return RenderHTML(w, v, code)
			}
		case "application/json", "application/*", "*/*":
			return RenderJSON(w, view, code)
		case "text/plain":
			return RenderPlain(w, view, code)
		}
	}
	return Abort(w, http.StatusNotAcceptable)
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
func RenderHTML(w http.ResponseWriter, view tmpl.Viewable, code int) error {
	b, err := renderer.Render(view)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, err = w.Write(b)
	return err
}

// RenderJSON writes the view as marshalled JSON.
func RenderJSON(w http.ResponseWriter, view Viewable, code int) error {
	b, err := json.Marshal(view)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if view == nil {
		return nil
	}
	_, err = w.Write(b)
	return err
}

// RenderPlain writes the view as a string.
func RenderPlain(w http.ResponseWriter, view Viewable, code int) error {
	s, ok := view.(string)
	if !ok {
		return fmt.Errorf("httpc: view for RenderPlain must be a string")
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, err := fmt.Fprintln(w, s)
	return err
}
