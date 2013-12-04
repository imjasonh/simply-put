package simplyput

// TODO: PropertyList can support nested objects with named properties like "A.B.C", support nested JSON objects.
// TODO: Add rudimentary single-property queries, pagination, sorting, etc.
// TODO: Add memcache
// TODO: Support ETags, If-Modified-Since, etc. (http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html)
// TODO: PUT requests
// TODO: HEAD requests
// TODO: PATCH requests/semantics
// TODO: Batch requests (via multipart?)
// TODO: User POSTs a JSON schema, future requests are validated against that schema. Would anybody use that?

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	idKey        = "_id"
	createdKey   = "_created"
	updatedKey   = "_updated"
	defaultLimit = 10
)

func init() {
	http.HandleFunc("/datastore/v1dev/objects/", handle)
}

type userQuery struct {
	Limit, Offset int
	FilterKey, FilterType, FilterValue,
	StartCursor, EndCursor string
}

// getUserID gets the Google User ID for an access token.
func getUserID(accessToken string, client http.Client) (string, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?access_token=" + accessToken)
	if err != nil {
		return "", err
	}
	var info struct {
		ID string
	}
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	resp.Body.Close()
	id := info.ID
	if id == "" {
		return "", errors.New("invalid auth")
	}
	return id, nil
}

// getKindAndID parses the kind and ID from a request path.
func getKindAndID(path string) (string, int64, error) {
	if match, err := regexp.MatchString("/datastore/v1dev/objects/[a-zA-Z]+/[0-9]+", path); err != nil {
		return "", int64(0), err
	} else if match {
		kind := path[len("/datastore/v1dev/objects/"):strings.LastIndex(path, "/")]
		idStr := path[strings.LastIndex(path, "/")+1:]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return "", int64(0), err
		}
		return kind, id, nil
	}
	if match, err := regexp.MatchString("/datastore/v1dev/objects/[a-zA-Z]+", path); err != nil {
		return "", int64(0), err
	} else if match {
		kind := path[len("/datastore/v1dev/objects/"):]
		return kind, int64(0), nil
	}
	return "", int64(0), errors.New("invalid path")
}

// handle dispatches requests to the relevant API method and arranges certain common state
func handle(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Add("Access-Control-Allow-Origin", "*")

	r.ParseForm()
	client := urlfetch.Client(c)

	var userID string
	if appengine.IsDevAppServer() {
		// For local development, don't require an access token or user ID
		// If the user_id param is set, that's the user ID.
		userID = r.Form.Get("user_id")
	} else {
		// Get the access_token from the request and turn it into a user ID with which we will namespace Kinds in the datastore.
		accessToken := r.Form.Get("access_token")
		if accessToken == "" {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				accessToken = h[len("Bearer "):]
			}
		}
		if accessToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		var err error // Needed because otherwise the next line shadows userID...
		userID, err = getUserID(accessToken, *client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	kind, id, err := getKindAndID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dsKind := fmt.Sprintf("%s--%s", userID, kind)

	resp := make(map[string]interface{}, 0)
	errCode := http.StatusOK
	if id == int64(0) {
		switch r.Method {
		case "POST":
			resp, errCode = insert(c, dsKind, r.Body)
			r.Body.Close()
		case "GET":
			resp, errCode = list(c, dsKind, newUserQuery(r))
		default:
			http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			return
		}
	} else {
		switch r.Method {
		case "GET":
			resp, errCode = get(c, dsKind, id)
		case "DELETE":
			errCode = delete(c, dsKind, id)
		case "POST":
			// This is strictly "replace all properties/values", not "add new properties, update existing"
			resp, errCode = update(c, dsKind, id, r.Body)
			r.Body.Close()
		default:
			http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			return
		}
	}
	if errCode != http.StatusOK {
		http.Error(w, "", errCode)
		return
	}
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
}

func newUserQuery(r *http.Request) userQuery {
	uq := userQuery{
		StartCursor: r.FormValue("start"),
		EndCursor:   r.FormValue("end"),
	}
	// TODO: MustParse for limit/offset, else panic
	uq.Limit, _ = strconv.Atoi(r.FormValue("limit"))
	uq.Offset, _ = strconv.Atoi(r.FormValue("offset"))

	// TODO: Support ?where=foo<bar queries (which may or may not be annoying to scope for users...)
	_ = r.FormValue("where")
	return uq
}

func delete(c appengine.Context, kind string, id int64) int {
	k := datastore.NewKey(c, kind, "", id, nil)
	if err := datastore.Delete(c, k); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return http.StatusNotFound
		} else {
			c.Errorf("%v", err)
			return http.StatusInternalServerError
		}
	}
	return http.StatusOK
}

func get(c appengine.Context, kind string, id int64) (map[string]interface{}, int) {
	k := datastore.NewKey(c, kind, "", id, nil)
	var pl plist
	if err := datastore.Get(c, k, &pl); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, http.StatusNotFound
		}
		c.Errorf("%v", err)
		return nil, http.StatusInternalServerError
	}
	m := plistToMap(pl, k.IntID())
	m[idKey] = k.IntID()
	return m, http.StatusOK
}

func insert(c appengine.Context, kind string, r io.Reader) (map[string]interface{}, int) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		c.Errorf("%v", err)
		return nil, http.StatusInternalServerError
	}
	m[createdKey] = time.Now().Unix()

	pl := mapToPlist("", m)

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err := datastore.Put(c, k, &pl)
	if err != nil {
		c.Errorf("%v", err)
		return nil, http.StatusInternalServerError
	}
	m[idKey] = k.IntID()
	return m, http.StatusOK
}

type prop struct {
	Name     string
	Value    interface{}
	Multiple bool
	NoIndex  bool
}
type plist []prop

// plistToMap transforms a plist such as you would get from the datastore into a map[string]interface{} suitable for JSON-encoding.
func plistToMap(pl plist, id int64) map[string]interface{} {
	m := make(map[string]interface{})
	for _, p := range pl {
		if strings.Contains(p.Name, ".") {
			parts := strings.Split(p.Name, ".")
			sub := m
			for _, p := range parts[:len(parts)-1] {
				// Traverse the path up until the leaf
				if i, exists := sub[p]; exists {
					// Already seen this path, traverse it
					if ii, ok := i.(map[string]interface{}); ok {
						sub = ii
					} else {
						// Got a sub-property of a non-map property. Uh oh...
						// Not sure it's worth failing/logging though...
					}
				} else {
					// First time down this path, add a new empty map
					next := map[string]interface{}{}
					sub[p] = next
					sub = next
				}
			}
			leaf := parts[len(parts)-1]
			if _, exists := sub[leaf]; exists {
				if !p.Multiple {
					// We would expect p.Multiple to be true here.
					// Not sure it's worth failing/logging though...
				}
				if _, isArr := sub[leaf].([]interface{}); isArr {
					// Already an array here, append to it
					sub[leaf] = append(sub[leaf].([]interface{}), p.Value)
				} else {
					// Already a single value here, should be an array now.
					sub[leaf] = []interface{}{sub[leaf], p.Value}
				}
			} else {
				sub[leaf] = p.Value
			}
		} else {
			if _, exists := m[p.Name]; exists {
				if !p.Multiple {
					// We would expect p.Multiple to be true here.
					// Not sure it's worth failing/logging though...
				}
				if _, isArr := m[p.Name].([]interface{}); isArr {
					// Already an array here, append to it
					m[p.Name] = append(m[p.Name].([]interface{}), p.Value)
				} else {
					// Already a single value here, should be an array now.
					m[p.Name] = []interface{}{m[p.Name], p.Value}
				}
			} else {
				m[p.Name] = p.Value
			}
		}
	}
	m[idKey] = id
	return m
}

// mapToPlist transforms a map[string]interface{} such as you would get from decoding JSON into a plist to store in the datastore.
func mapToPlist(prefix string, m map[string]interface{}) plist {
	pl := make(plist, 0, len(m))
	for k, v := range m {
		if m, nest := v.(map[string]interface{}); nest {
			// Generate a plist for this sub-map, and append it
			pl = append(pl, mapToPlist(prefix+k+".", m)...)
		} else if _, mult := v.([]interface{}); mult {
			// Generate a prop for every item in the slice
			for _, mv := range v.([]interface{}) {
				pl = append(pl, prop{
					Name:     prefix + k,
					Value:    mv,
					Multiple: true,
				})
			}
			// TODO: Apparently no way to store an empty list? That seems odd...
		} else {
			pl = append(pl, prop{
				Name:  prefix + k,
				Value: v,
			})
		}
	}
	return pl
}

func list(c appengine.Context, kind string, uq userQuery) (map[string]interface{}, int) {
	q := datastore.NewQuery(kind)

	if uq.Limit != 0 {
		q = q.Limit(uq.Limit)
	}
	if c, err := datastore.DecodeCursor(uq.StartCursor); err == nil {
		q.Start(c)
	}
	if c, err := datastore.DecodeCursor(uq.EndCursor); err == nil {
		q.End(c)
	}

	items := make([]map[string]interface{}, 0)

	var crs datastore.Cursor
	for t := q.Run(c); ; {
		var pl plist
		k, err := t.Next(&pl)
		if err == datastore.Done {
			break
		}
		if err != nil {
			c.Errorf("%v", err)
			return nil, http.StatusInternalServerError
		}
		m := plistToMap(pl, k.IntID())
		items = append(items, m)
		if crs, err = t.Cursor(); err != nil {
			c.Errorf("%v", err)
			return nil, http.StatusInternalServerError
		}
	}
	r := map[string]interface{}{
		"items":          items,
		"nextStartToken": crs.String(),
	}
	return r, http.StatusOK
}

func update(c appengine.Context, kind string, id int64, r io.Reader) (map[string]interface{}, int) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		c.Errorf("%v", err)
		return nil, http.StatusInternalServerError
	}
	m[updatedKey] = time.Now().Unix()

	pl := mapToPlist("", m)

	k := datastore.NewKey(c, kind, "", id, nil)
	if _, err := datastore.Put(c, k, &pl); err != nil {
		c.Errorf("%v", err)
		return nil, http.StatusInternalServerError
	}
	m[idKey] = id
	return m, http.StatusOK
}
