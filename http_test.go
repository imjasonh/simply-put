package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/boltdb/bolt"
)

const testFile = "test.bolt.db"

// TODO: Extend this test by making multiple requests and recording responses
func TestCreateUpdateDelete(t *testing.T) {
	for _, c := range []struct {
		reqs []req
	}{{
		// Insert an entity then retrieve it. Delete it and get a 404.
		[]req{
			{"POST", "/MyKind", `{"a":true}`, http.StatusOK,
				map[string]interface{}{"a": true}},
			{"GET", "/MyKind/foo", "", http.StatusOK,
				map[string]interface{}{"a": true}},
			{"POST", "/MyKind/foo", `{"a":false}`, http.StatusOK,
				map[string]interface{}{"a": false}},
			{"GET", "/MyKind/foo", "", http.StatusOK,
				map[string]interface{}{"a": false}},
			{"DELETE", "/MyKind/foo", "", http.StatusOK, nil},
			{"GET", "/MyKind/foo", "", http.StatusNotFound, nil},
		},
	}} {
		os.Remove(testFile)
		db, err := bolt.Open(testFile, 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		defer os.Remove(testFile)
		s := &server{db}
		t.Logf("server up")
		for _, r := range c.reqs {
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

			s.ServeHTTP(w, req)
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

			// TODO: only tests that expResp is a subset of gotResp, doesn't test full equality
			//       because _id is a random number.
			for k, v := range r.expResp {
				if gotResp[k] != v {
					t.Errorf("unexpected result:\n got %v\nwant %v", gotResp, r.expResp)
					break
				}
			}
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
