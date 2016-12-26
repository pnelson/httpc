package httpc

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

// Viewable represents a view. To provide an expressive API, this
// type is an alias for interface{} that is named for documentation.
type Viewable interface{}

// Renderable represents the ability to render HTML templates.
type Renderable interface {
	Render(view interface{}) ([]byte, error)
}

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
			v, ok := view.(Renderable)
			if !ok {
				continue
			}
			return RenderHTML(w, v, code)
		case "application/json", "application/*", "*/*":
			return RenderJSON(w, view, code)
		case "text/plain":
			return RenderPlain(w, view, code)
		}
	}
	return Abort(w, http.StatusNotAcceptable)
}

// RenderHTML writes the view as templated HTML.
func RenderHTML(w http.ResponseWriter, view Renderable, code int) error {
	b, err := view.Render(view)
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
