package httpc

import (
	"encoding/json"
	"mime"
	"net/http"

	"github.com/gorilla/schema"
)

// A Form represents a form with validation.
type Form interface {
	// Validate sanitizes and validates the form.
	Validate() error
}

// UploadForm represents a form with a maximum file upload size.
type UploadForm interface {
	// MaxUploadSize returns the maximum file upload size in bytes.
	MaxUploadSize() int64
}

// Validate decodes, sanitizes and validates the request body
// and stores the result in to the value pointed to by form.
func Validate(req *http.Request, form Form) error {
	v := req.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(v)
	if err != nil {
		return err
	}
	switch media {
	case "application/json":
		return ValidateJSON(req, form)
	case "multipart/form-data":
		return ValidateMultipart(req, form)
	}
	return ValidateForm(req, form)
}

// decoder decodes a struct with form values.
// The decoder caches struct meta data and can be shared safely.
var decoder = schema.NewDecoder()

// ValidateForm decodes, sanitizes and validates the request
// body as a form and stores the result in the value pointed
// to by form.
func ValidateForm(req *http.Request, form Form) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}
	err = decoder.Decode(form, req.PostForm)
	if err != nil {
		return err
	}
	return form.Validate()
}

// ValidateJSON decodes, sanitizes and validates the request
// body as JSON and stores the result in the value pointed
// to by form.
func ValidateJSON(req *http.Request, form Form) error {
	defer req.Body.Close()
	err := json.NewDecoder(req.Body).Decode(form)
	if err != nil {
		return err
	}
	return form.Validate()
}

// DefaultMaxUploadSize is the default maximum file upload size in bytes.
const DefaultMaxUploadSize int64 = 32 << 20 // 32 MB

// ValidateMultipart decodes, sanitizes and validates the request
// body as multipart/form-data and stores the result in the value
// pointed to by form.
func ValidateMultipart(req *http.Request, form Form) error {
	maxUploadSize := DefaultMaxUploadSize
	uf, ok := form.(UploadForm)
	if ok {
		maxUploadSize = uf.MaxUploadSize()
	}
	err := req.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return err
	}
	err = decoder.Decode(form, req.MultipartForm.Value)
	if err != nil {
		return err
	}
	return form.Validate()
}
