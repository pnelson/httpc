// Package httpc implements HTTP request and response helpers.
package httpc

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"
)

// Context represents the request context.
type Context struct {
	context.Context
	http.ResponseWriter
	Request *http.Request
}

// key represents httpc context.Context keys.
type key int

// Package context.Context keys.
const keyError key = iota

// mu protects variables that Context uses but are not
// expected to change beyond application initialization.
var mu sync.Mutex

// NewContext returns a new request context.
func NewContext(ctx context.Context, w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Context:        ctx,
		ResponseWriter: w,
		Request:        req,
	}
}

// Abort replies to the request with a default plain text error.
func Abort(w http.ResponseWriter, code int) error {
	return RenderPlain(w, StatusText(code), code)
}

// Abort replies to the request with a default plain text error.
func (ctx *Context) Abort(code int) error {
	return Abort(ctx, code)
}

// NoContent writes http.StatusNoContent to the header.
func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// NoContent writes http.StatusNoContent to the header.
func (ctx *Context) NoContent() error {
	return NoContent(ctx)
}

// Redirect replies to the request with a redirect to path.
// This is the equivalent to http.Redirect and is here for consistency.
func Redirect(w http.ResponseWriter, req *http.Request, path string, code int) error {
	http.Redirect(w, req, path, code)
	return nil
}

// Redirect replies to the request with a redirect to path.
func (ctx *Context) Redirect(path string, code int) error {
	return Redirect(ctx, ctx.Request, path, code)
}

// RedirectTo replies to the request with a redirect to the application
// path constructed from the format specifier and args.
func RedirectTo(w http.ResponseWriter, req *http.Request, format string, args ...interface{}) error {
	return Redirect(w, req, fmt.Sprintf(format, args...), http.StatusSeeOther)
}

// RedirectTo replies to the request with a redirect to the application
// path constructed from the format specifier and args.
func (ctx *Context) RedirectTo(format string, args ...interface{}) error {
	return RedirectTo(ctx, ctx.Request, format, args)
}

// RemoteAddr returns a best guess remote address.
func RemoteAddr(req *http.Request) string {
	addr := req.Header.Get("X-Real-IP")
	if len(addr) == 0 {
		addr = req.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = req.RemoteAddr
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return addr
			}
			return host
		}
	}
	return addr
}

// RemoteAddr returns a best guess remote address.
func (ctx *Context) RemoteAddr() string {
	return RemoteAddr(ctx.Request)
}

// SetCookie adds a Set-Cookie header to the provided
// http.ResponseWriter's headers. The provided cookie must
// have a valid Name. Invalid cookies may be silently dropped.
func SetCookie(w http.ResponseWriter, cookie *http.Cookie) {
	if cookie.MaxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
	} else {
		cookie.Expires = time.Unix(1, 0)
	}
	http.SetCookie(w, cookie)
}

// SetCookie adds a Set-Cookie header to the provided
// http.ResponseWriter's headers. The provided cookie must
// have a valid Name. Invalid cookies may be silently dropped.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	SetCookie(ctx, cookie)
}

// Error returns the error response if any.
func (ctx *Context) Error() error {
	err, ok := ctx.Value(keyError).(error)
	if !ok {
		return nil
	}
	return err
}

// setError sets the error on the context.
func (ctx *Context) setError(err error) {
	ctx.Context = context.WithValue(ctx, keyError, err)
}
