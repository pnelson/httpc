package httpc

import (
	"net/http"
	"testing"
)

func TestStatusText(t *testing.T) {
	tests := []int{
		StatusMulti,
		StatusUnprocessableEntity,
		StatusLocked,
		StatusFailedDependency,
		StatusInsufficientStorage,
		http.StatusOK,
	}
	for i, code := range tests {
		have := StatusText(code)
		if have == "" {
			t.Errorf("%d. code %d should return status text", i, code)
		}
	}
}
