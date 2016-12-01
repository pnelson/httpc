package httpc

import (
	"context"
	"net/http"

	"goji.io"
	"goji.io/middleware"
	"goji.io/pat"
	"goji.io/pattern"
)

// Mux represents an HTTP request multiplexer.
type Mux struct {
	*goji.Mux
	errorHandler http.Handler
}

// Handler represents a HTTP handler with error handling.
type Handler func(w http.ResponseWriter, req *http.Request) error

// NewMux returns a new mux.
func NewMux() *Mux {
	return &Mux{
		Mux:          goji.NewMux(),
		errorHandler: http.HandlerFunc(defaultErrorHandler),
	}
}

// NewSubMux returns a new mux mounted at the given pattern p.
func (m *Mux) NewSubMux(p string) *Mux {
	h := &Mux{Mux: goji.SubMux()}
	m.Handle(p, h)
	return h
}

// Any registers a route that matches any HTTP method.
func (m *Mux) Any(p string, h Handler) {
	m.handle(pat.New(p), h)
}

// Delete registers a route that only matches the DELETE HTTP method.
func (m *Mux) Delete(p string, h Handler) {
	m.handle(pat.Delete(p), h)
}

// Get registers a route that only matches the GET and HEAD HTTP methods.
// HEAD requests are handled transparently by net/http.
func (m *Mux) Get(p string, h Handler) {
	m.handle(pat.Get(p), h)
}

// Head registers a route that only matches the HEAD HTTP method.
func (m *Mux) Head(p string, h Handler) {
	m.handle(pat.Head(p), h)
}

// Options registers a route that only matches the OPTIONS HTTP method.
func (m *Mux) Options(p string, h Handler) {
	m.handle(pat.Options(p), h)
}

// Patch registers a route that only matches the PATCH HTTP method.
func (m *Mux) Patch(p string, h Handler) {
	m.handle(pat.Patch(p), h)
}

// Post registers a route that only matches the POST HTTP method.
func (m *Mux) Post(p string, h Handler) {
	m.handle(pat.Post(p), h)
}

// Put registers a route that only matches the PUT HTTP method.
func (m *Mux) Put(p string, h Handler) {
	m.handle(pat.Put(p), h)
}

// handle registers a route with the mux.
func (m *Mux) handle(p *pat.Pattern, h Handler) {
	fn := func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err != nil {
			ctx := req.Context()
			ctx = context.WithValue(ctx, keyError, err)
			req = req.WithContext(ctx)
			m.errorHandler.ServeHTTP(w, req)
		}
	}
	m.HandleFunc(p, fn)
}

// Handle registers a standard net/http route with the mux.
func (m *Mux) Handle(p string, h http.Handler) {
	m.Mux.Handle(pat.New(p), h)
}

// FileServer registers a file system with the mux.
// The pattern p is expected to be a prefix wildcard route.
// See https://godoc.org/goji.io/pat#hdr-Prefix_Matches.
// The pattern prefix is removed from the request URL before handled.
func (m *Mux) FileServer(p string, fs http.FileSystem) {
	prefix := p[:len(p)-1]
	m.Handle(p, http.StripPrefix(prefix, http.FileServer(fs)))
}

// SetErrorHandler sets the http.Handler to delegate
// to when errors are returned.
func (m *Mux) SetErrorHandler(h http.Handler) {
	m.errorHandler = h
}

// Error returns the error response if any.
func Error(req *http.Request) error {
	err, ok := req.Context().Value(keyError).(error)
	if !ok {
		return nil
	}
	return err
}

// MatchedHandler returns the handler corresponding to the most
// recently matched pattern, or nil if no pattern was matched.
func MatchedHandler(req *http.Request) http.Handler {
	return middleware.Handler(req.Context())
}

// Path returns the escaped request path.
func Path(req *http.Request) string {
	return pattern.Path(req.Context())
}

// Param returns the bound parameter with the given name.
func Param(req *http.Request, name string) string {
	return pat.Param(req, name)
}

// Query returns the first query value associated with the given key.
// If there are no values associated with the key, Query returns the
// empty string.
func Query(req *http.Request, name string) string {
	return req.URL.Query().Get(name)
}

// defaultErrorHandler is the default error handler.
func defaultErrorHandler(w http.ResponseWriter, req *http.Request) {
	Abort(w, http.StatusInternalServerError)
}
