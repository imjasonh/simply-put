package simplyput

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"appengine/aetest"
)

// TODO: Extend this test by making multiple requests and recording responses
func TestFoo(t *testing.T) {
	now := time.Unix(0, 0)
	nowFunc = func() time.Time {
		return now
	}
	for _, c := range []struct {
		reqs []req
	}{{
		// Insert an entity then retrieve it. Delete it and get a 404.
		[]req{
			{"POST", "/MyKind", `{"a":true}`, http.StatusOK,
				map[string]interface{}{"a": true, "_created": float64(0)}},
			{"GET", "/MyKind/{{ID}}", "", http.StatusOK,
				map[string]interface{}{"a": true, "_created": float64(0)}},
			{"POST", "/MyKind/{{ID}}", `{"a":false}`, http.StatusOK,
				map[string]interface{}{"a": false, "_updated": float64(2)}}, // updated 2s after created
			{"GET", "/MyKind/{{ID}}", "", http.StatusOK,
				map[string]interface{}{"a": false, "_updated": float64(2)}},
			{"DELETE", "/MyKind/{{ID}}", "", http.StatusOK, nil},
			{"GET", "/MyKind/{{ID}}", "", http.StatusNotFound, nil},
		},
	}} {
		ctx, err := aetest.NewContext(nil)
		if err != nil {
			t.Errorf("unexpected error creating context: %v", err)
			continue
		}
		defer ctx.Close()

		var lastID int64
		for _, r := range c.reqs {
			t.Logf("now %v", now)
			r.path = strings.Replace(r.path, "{{ID}}", fmt.Sprintf("%d", lastID), -1)
			var body io.Reader
			if r.body != "" {
				body = strings.NewReader(r.body)
			}
			t.Logf("request: %v", r)

			req, err := http.NewRequest(r.method, r.path, body)
			if err != nil {
				t.Errorf("unexpected error creating request: %v", err)
				continue
			}
			w := httptest.NewRecorder()
			handle(ctx, w, req)
			t.Logf("response: %d\n%s", w.Code, string(w.Body.Bytes()))

			if got, want := w.Code, r.expCode; got != want {
				t.Errorf("unexpected code: got %d, want %d", got, want)
			}
			if r.expResp == nil {
				// TODO: For DELETE, response is empty. For GET (404), response is "\n"
				if w.Body.Len() > 1 {
					t.Errorf("unexpectedly got response: %q", w.Body.String())
				}
				continue
			}

			var gotResp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&gotResp); err != nil {
				t.Errorf("unexpected error decoding response: %v", err)
				continue
			}

			lastID = int64(gotResp["_id"].(float64))
			// TODO: only tests that expResp is a subset of gotResp, doesn't test full equality
			//       because _id is a random number.
			for k, v := range r.expResp {
				if gotResp[k] != v {
					t.Errorf("unexpected result:\n got %v\nwant %v", gotResp, r.expResp)
					break
				}
			}
			now = now.Add(time.Second)
		}
	}
}

type req struct {
	method  string
	path    string
	body    string                 // if "", no body provided
	expCode int                    // if 0, StatusOK expected
	expResp map[string]interface{} // If nil, no response body expected
}
