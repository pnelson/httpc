// Package httpc implements HTTP request and response helpers.
package httpc

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// key represents httpc context.Context keys.
type key int

// Package context.Context keys.
const keyError key = iota

// Abort replies to the request with a default plain text error.
func Abort(w http.ResponseWriter, code int) error {
	return RenderPlain(w, http.StatusText(code), code)
}

// NoContent writes http.StatusNoContent to the header.
func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// Redirect replies to the request with a redirect to path.
// This is the equivalent to http.Redirect and is here for consistency.
func Redirect(w http.ResponseWriter, req *http.Request, path string, code int) error {
	http.Redirect(w, req, path, code)
	return nil
}

// RedirectTo replies to the request with a redirect to the application
// path constructed from the format specifier and args.
func RedirectTo(w http.ResponseWriter, req *http.Request, format string, args ...interface{}) error {
	return Redirect(w, req, fmt.Sprintf(format, args...), http.StatusSeeOther)
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

// SetCookie adds a Set-Cookie header to the provided
// http.ResponseWriter's headers. The provided cookie must
// have a valid Name. Invalid cookies may be silently dropped.
func SetCookie(w http.ResponseWriter, cookie *http.Cookie) {
	if cookie.MaxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
	} else if cookie.MaxAge < 0 {
		cookie.Expires = time.Unix(1, 0)
	}
	http.SetCookie(w, cookie)
}
