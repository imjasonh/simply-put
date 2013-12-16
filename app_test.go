package simplyput

import (
	"reflect"
	"testing"
)

const id = int64(123)

func TestThereAndBackAgain(t *testing.T) {
	cases := []struct {
		pl plist
		m  map[string]interface{}
	}{{
		// Empty plist -> nearly-empty map
		[]prop{},
		map[string]interface{}{
			"_id": id,
		},
	}, {
		// Single-property plist -> single-property map
		[]prop{{
			Name:  "foo",
			Value: "bar",
		}},
		map[string]interface{}{
			"foo": "bar",
			"_id": id,
		},
	}, {
		// Lists of arbitrary types
		[]prop{{
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
		[]prop{{
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
		[]prop{{
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
