package httpc

import (
	"encoding/json"
	"mime"
	"net/http"
	"reflect"

	"github.com/gorilla/schema"
)

// A Form represents a form with validation.
type Form interface {
	// Validate sanitizes and validates the form.
	Validate() error
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

// Validate decodes, sanitizes and validates the request body
// and stores the result in to the value pointed to by form.
func (ctx *Context) Validate(form Form) error {
	return Validate(ctx.Request, form)
}

// validate validates nested fields if the form is a struct
// and then validates itself.
func validate(form Form) error {
	v := reflect.ValueOf(form)
	t := v.Type()
	err := validateFields(v, t)
	if err != nil {
		return err
	}
	return form.Validate()
}

// validateFields validates the fields of struct if applicable.
// Embedded fields are validated as if they are at the same level.
// Since we need the type assertion back to Form, the embedded field
// must be a pointer or exported.
func validateFields(v reflect.Value, t reflect.Type) error {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		t = v.Type()
		for i := 0; i < t.NumField(); i++ {
			fieldVal := v.Field(i)
			fieldType := t.Field(i)
			if fieldType.Anonymous {
				err := validateFields(fieldVal, fieldType.Type)
				if err != nil {
					return err
				}
			}
			err := validateField(fieldVal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// validateField validates a struct field.
func validateField(v reflect.Value) error {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	if !v.CanInterface() {
		return nil
	}
	form, ok := v.Interface().(Form)
	if !ok {
		return nil
	}
	return validate(form)
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
	return validate(form)
}

// ValidateForm decodes, sanitizes and validates the request
// body as a form and stores the result in the value pointed
// to by form.
func (ctx *Context) ValidateForm(form Form) error {
	return ValidateForm(ctx.Request, form)
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
	return validate(form)
}

// ValidateJSON decodes, sanitizes and validates the request
// body as JSON and stores the result in the value pointed
// to by form.
func (ctx *Context) ValidateJSON(form Form) error {
	return ValidateJSON(ctx.Request, form)
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
func ValidateMultipart(req *http.Request, form Form) error {
	err := req.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return err
	}
	err = decoder.Decode(form, req.MultipartForm.Value)
	if err != nil {
		return err
	}
	return validate(form)
}

// ValidateMultipart decodes, sanitizes and validates the request
// body as multipart/form-data and stores the result in the value
// pointed to by form.
func (ctx *Context) ValidateMultipart(form Form) error {
	return ValidateMultipart(ctx.Request, form)
}
