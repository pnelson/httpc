package httpc

import (
	"net/http"

	"goji.io"
	"goji.io/pat"
	"goji.io/pattern"
	"golang.org/x/net/context"
)

// Mux represents an HTTP request multiplexer.
type Mux struct {
	*goji.Mux
	errorHandler Handler
}

// Handler represents a HTTP handler.
type Handler func(*Context) error

// NewMux returns a new mux.
func NewMux() *Mux {
	return &Mux{Mux: goji.NewMux(), errorHandler: defaultErrorHandler}
}

// NewSubMux returns a new mux mounted at the given pattern p.
func (m *Mux) NewSubMux(p string) *Mux {
	h := &Mux{Mux: goji.SubMux()}
	m.HandleC(pat.New(p), h)
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
	fn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		c := NewContext(ctx, w, req)
		err := h(c)
		if err != nil {
			c.setError(err)
			m.errorHandler(c)
		}
	}
	m.HandleFuncC(p, fn)
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

// SetErrorHandler sets the Handler to delegate to when errors are returned.
func (m *Mux) SetErrorHandler(h Handler) {
	m.errorHandler = h
}

// ServeHTTP implements the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	m.ServeHTTPC(ctx, w, req)
}

// Path returns the escaped request path.
func (ctx *Context) Path() string {
	return pattern.Path(ctx)
}

// Param returns the bound parameter with the given name.
func (ctx *Context) Param(name string) string {
	return pat.Param(ctx, name)
}

// defaultErrorHandler is the default error handler.
func defaultErrorHandler(ctx *Context) error {
	return ctx.Abort(http.StatusInternalServerError)
}
