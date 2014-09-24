package simplyput

import (
	"net/http"
	"strings"
	"testing"

	"appengine/aetest"
)

// TODO: Extend this test by making multiple requests and recording responses
func TestFoo(t *testing.T) {
	for _, c := range []struct {
		kind string
		body string
	}{{
		"123-Thing",
		"{}",
	}} {
		ctx, err := aetest.NewContext(nil)
		if err != nil {
			t.Errorf("unexpected error creating context: %v", err)
			continue
		}
		_, code := insert(ctx, c.kind, strings.NewReader(c.body))
		if code != http.StatusOK {
			t.Errorf("unexpected code: %d", code)
		}
	}
}
