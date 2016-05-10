package httpc

import (
	"encoding/json"
	"mime"

	"github.com/gorilla/schema"
)

// A Form represents a form with validation.
type Form interface {
	// Validate sanitizes and validates the form.
	Validate() error
}

// Validate decodes, sanitizes and validates the request body
// and stores the result in to the value pointed to by form.
func (ctx *Context) Validate(form Form) error {
	v := ctx.Request.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(v)
	if err != nil {
		return err
	}
	switch media {
	case "application/json":
		return ctx.ValidateJSON(form)
	case "multipart/form-data":
		return ctx.ValidateMultipart(form)
	}
	return ctx.ValidateForm(form)
}

// decoder decodes a struct with form values.
// The decoder caches struct meta data and can be shared safely.
var decoder = schema.NewDecoder()

// ValidateForm decodes, sanitizes and validates the request
// body as a form and stores the result in the value pointed
// to by form.
func (ctx *Context) ValidateForm(form Form) error {
	err := ctx.Request.ParseForm()
	if err != nil {
		return err
	}
	err = decoder.Decode(form, ctx.Request.PostForm)
	if err != nil {
		return err
	}
	return form.Validate()
}

// ValidateJSON decodes, sanitizes and validates the request
// body as JSON and stores the result in the value pointed
// to by form.
func (ctx *Context) ValidateJSON(form Form) error {
	defer ctx.Request.Body.Close()
	err := json.NewDecoder(ctx.Request.Body).Decode(form)
	if err != nil {
		return err
	}
	return form.Validate()
}

// DefaultMaxUploadSize is the default maximum file upload size in bytes.
const DefaultMaxUploadSize int64 = 32 << 20 // 32 MB

// maxUploadSize is the maximum file upload size in bytes.
var maxUploadSize = DefaultMaxUploadSize

// SetMaxUploadSize sets the maximum file upload size in bytes.
func SetMaxUploadSize(size int64) {
	mu.Lock()
	maxUploadSize = size
	mu.Unlock()
}

// ValidateMultipart decodes, sanitizes and validates the request
// body as multipart/form-data and stores the result in the value
// pointed to by form.
func (ctx *Context) ValidateMultipart(form Form) error {
	err := ctx.Request.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return err
	}
	err = decoder.Decode(form, ctx.Request.MultipartForm.Value)
	if err != nil {
		return err
	}
	return form.Validate()
}
