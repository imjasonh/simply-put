package main

import (
	"net/http"
	"reflect"
	"testing"
)

func TestUserQuery(t *testing.T) {
	cases := []struct {
		r        http.Request
		uq       *userQuery
		hasError bool
	}{{
		// User didn't specify any params
		http.Request{},
		&userQuery{Limit: defaultLimit},
		false,
	}, {
		// User requests all the params
		http.Request{
			Form: map[string][]string{
				"limit": []string{"1"},
				"start": []string{"s"},
				"end":   []string{"e"},
				"sort":  []string{"-foo"},
				"where": []string{"foo=bar", "baz=qux", "quux=duck"},
			},
		},
		&userQuery{Limit: 1, StartCursor: "s", EndCursor: "e", Sort: "-foo", Filters: []filter{
			{Key: "foo", Value: "bar"},
			{Key: "baz", Value: "qux"},
			{Key: "quux", Value: "duck"},
		}},
		false,
	}, {
		// User passes non-numerical "limit" param
		http.Request{
			Form: map[string][]string{
				"limit": []string{"bad"},
			},
		},
		nil,
		true,
	}, {
		// User passes malformed "where" param
		http.Request{
			Form: map[string][]string{
				"where": []string{"bad"},
			},
		},
		nil,
		true,
	}}
	for _, c := range cases {
		a, err := newUserQuery(&c.r)
		if c.hasError && err == nil {
			t.Errorf("newUserQuery(%v); expected error", c.r)
		} else if err != nil && !c.hasError {
			t.Errorf("unexpected error %v", err)
		} else if !reflect.DeepEqual(c.uq, a) {
			t.Errorf("newUserQuery(%v);\n got %#v\nwant %#v", c.r, a, c.uq)
		}
	}
}

func TestGetKindAndID(t *testing.T) {
	cases := []struct {
		path     string
		kind, id string
		hasError bool
	}{
		{"/MyKindOfData", "MyKindOfData", "", false},
		{"/MyKindOfData/foo", "MyKindOfData", "foo", false},

		{"/bad/path/too/long", "", "", true},
		{"bad/path", "", "", true},
		{"/", "", "", true},
	}
	for _, c := range cases {
		kind, id, err := getKindAndID(c.path)
		if c.hasError && err == nil {
			t.Errorf("getKindAndID(%s); expected error, got %s,%s", c.path, kind, id)
		} else if err != nil && !c.hasError {
			t.Errorf("unexpected error %v", err)
		} else if c.kind != kind || c.id != id {
			t.Errorf("getKindAndID(%s); got %s,%s want %s,%s", c.path, kind, id, c.kind, c.id)
		}
	}
}
