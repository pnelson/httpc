package httpc

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testForm struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func (f testForm) Validate() error {
	if f.Bar < 1 {
		return errors.New("f.Bar < 1")
	}
	return nil
}

func TestValidateJSON(t *testing.T) {
	tests := map[string]struct {
		body    string
		isValid bool
	}{
		"valid":   {`{"foo":"bar","bar":1}`, true},
		"invalid": {`{"foo":"bar","bar":0}`, false},
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
