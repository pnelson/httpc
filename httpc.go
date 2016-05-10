// Package httpc implements HTTP request and response helpers.
package httpc

import (
	"fmt"
	"net/http"
	"strings"
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
func (ctx *Context) Abort(code int) {
	ctx.RenderPlain(StatusText(code), code)
}

// NoContent writes http.StatusNoContent to the header.
func (ctx *Context) NoContent() error {
	ctx.WriteHeader(http.StatusNoContent)
	return nil
}

// Redirect replies to the request with a redirect to path.
func (ctx *Context) Redirect(path string, code int) error {
	http.Redirect(ctx, ctx.Request, path, code)
	return nil
}

// RedirectTo replies to the request with a redirect to the application
// path constructed from the format specifier and args.
func (ctx *Context) RedirectTo(format string, args ...interface{}) error {
	return ctx.Redirect(fmt.Sprintf(format, args...), http.StatusSeeOther)
}

// RemoteAddr returns a best guess remote address.
func (ctx *Context) RemoteAddr() string {
	addr := ctx.Request.Header.Get("X-Real-IP")
	if len(addr) == 0 {
		addr = ctx.Request.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = ctx.Request.RemoteAddr
			if i := strings.LastIndex(addr, ":"); i > -1 {
				addr = addr[:i]
			}
		}
	}
	return addr
}

// SetCookie adds a Set-Cookie header to the provided
// http.ResponseWriter's headers. The provided cookie must
// have a valid Name. Invalid cookies may be silently dropped.
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	if cookie.MaxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
	} else {
		cookie.Expires = time.Unix(1, 0)
	}
	http.SetCookie(ctx, cookie)
}
