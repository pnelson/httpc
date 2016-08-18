package httpc

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testForm struct {
	Foo int       `json:"foo"`
	Bar testField `json:"bar"`
	*testEmbedded
}

func (f testForm) Validate() error {
	if f.Foo < 1 {
		return errors.New("f.Foo < 1")
	}
	return nil
}

type testField int

func (f testField) Validate() error {
	if f < 1 {
		return errors.New("testField < 1")
	}
	return nil
}

type testEmbedded struct {
	Baz int       `json:"baz"`
	Qux *testForm `json:"qux,omitempty"`
}

func (f testEmbedded) Validate() error {
	if f.Baz < 1 {
		return errors.New("f.Baz < 1")
	}
	return nil
}

func TestValidateJSON(t *testing.T) {
	tests := map[string]struct {
		body    string
		isValid bool
	}{
		"valid":                   {`{"foo":1,"bar":1,"baz":1}`, true},
		"invalid":                 {`{"foo":0,"bar":1,"baz":1}`, false},
		"field invalid":           {`{"foo":1,"bar":0,"baz":1}`, false},
		"embedded invalid":        {`{"foo":1,"bar":1,"baz":0}`, false},
		"nested valid":            {`{"foo":1,"bar":1,"baz":1,"qux":{"foo":1,"bar":1,"baz":1}}`, true},
		"nested invalid":          {`{"foo":1,"bar":1,"baz":1,"qux":{"foo":0,"bar":1,"baz":1}}`, false},
		"nested field invalid":    {`{"foo":1,"bar":1,"baz":1,"qux":{"foo":1,"bar":0,"baz":1}}`, false},
		"nested embedded invalid": {`{"foo":1,"bar":1,"baz":1,"qux":{"foo":1,"bar":1,"baz":0}}`, false},
	}
	for name, tt := range tests {
		var form testForm
		req := testRequest(t, strings.NewReader(tt.body))
		err := ValidateJSON(req, &form)
		switch {
		case tt.isValid && err != nil:
			t.Errorf("TestValidateJSON %s: %v", name, err)
		case !tt.isValid && err == nil:
			t.Errorf("TestValidateJSON %s: expected error", name)
		}
	}
}

func testRequest(t *testing.T, body io.Reader) *http.Request {
	req, err := http.NewRequest(http.MethodPost, "http://localhost", body)
	if err != nil {
		t.Fatal(err)
	}
	return req
}
