package simplyput

import (
	"net/http"
	"reflect"
	"testing"

	"appengine/datastore"
)

const id = int64(123)

func TestThereAndBackAgain(t *testing.T) {
	cases := []struct {
		pl datastore.PropertyList
		m  map[string]interface{}
	}{{
		// Empty plist -> nearly-empty map
		[]datastore.Property{},
		map[string]interface{}{
			"_id": id,
		},
	}, {
		// Single-property plist -> single-property map
		[]datastore.Property{{
			Name:  "foo",
			Value: "bar",
		}},
		map[string]interface{}{
			"foo": "bar",
			"_id": id,
		},
	}, {
		// Lists of arbitrary types
		[]datastore.Property{{
			Name:     "foo",
			Value:    "a",
			Multiple: true,
		}, {
			Name:     "foo",
			Value:    1,
			Multiple: true,
		}, {
			Name:     "foo",
			Value:    true,
			Multiple: true,
		}},
		map[string]interface{}{
			"foo": []interface{}{"a", 1, true},
			"_id": id,
		},
	}, {
		// Nested properties -> nested maps
		[]datastore.Property{{
			Name:  "a.b.c",
			Value: true,
		}},
		map[string]interface{}{
			"_id": id,
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": true,
				},
			},
		},
	}, {
		// Nested properties with a list at the leaf
		[]datastore.Property{{
			Name:     "a.b.c",
			Value:    true,
			Multiple: true,
		}, {
			Name:     "a.b.c",
			Value:    1,
			Multiple: true,
		}},
		map[string]interface{}{
			"_id": id,
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": []interface{}{true, 1},
				},
			},
		},
	}}
	for _, c := range cases {
		m := plistToMap(c.pl, id)
		if !reflect.DeepEqual(c.m, m) {
			t.Errorf("plistToMap(%v, %d); got %#v want %#v", c.pl, id, c.m, m)
		}

		delete(m, "_id")
		pl := mapToPlist("", m)
		if !reflect.DeepEqual(c.pl, pl) {
			t.Errorf("mapToPlist(%v); got %#v want %#v", m, pl, c.pl)
		}
	}
}

func TestUserQuery(t *testing.T) {
	cases := []struct {
		r        http.Request
		uq       *userQuery
		hasError bool
	}{{
		// User didn't specify any params
		http.Request{},
		&userQuery{},
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
			t.Errorf("newUserQuery(%v); got %#v want %#v", c.r, a, c.uq)
		}
	}
}

func TestGetKindAndID(t *testing.T) {
	cases := []struct {
		path     string
		kind     string
		id       int64
		hasError bool
	}{
		{"/MyKindOfData", "MyKindOfData", 0, false},
		{"/MyKindOfData/123", "MyKindOfData", 123, false},
		{"/123", "123", 0, false}, // Not sure if this is actually valid...

		{"/bad/path/too/long", "", 0, true},
		{"bad/path", "", 0, true},
		{"/MyKindOfData/badid", "", 0, true},
		{"/", "", 0, true},
	}
	for _, c := range cases {
		kind, id, err := getKindAndID(c.path)
		if c.hasError && err == nil {
			t.Errorf("getKindAndID(%s); expected error, got %#s,%d", c.path, kind, id)
		} else if err != nil && !c.hasError {
			t.Errorf("unexpected error %v", err)
		} else if c.kind != kind || c.id != id {
			t.Errorf("getKindAndID(%s); got %s,%d want %s,%d", c.path, kind, id, c.kind, c.id)
		}
	}
}
