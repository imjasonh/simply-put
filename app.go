package datastoreapi

// TODO: Figure out if PropertyList can support nested objects, or fail if they are detected.
// TODO: Add rudimentary single-property queries, pagination, sorting, etc.

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
	http.HandleFunc("/datastore/v1dev/objects/", datastoreAPI)
}

type UserQuery struct {
	Limit, Offset int
	FilterKey, FilterType, FilterValue,
	StartCursor, EndCursor string
}

type UserInfo struct {
	ID string
}

// getUserID gets the Google User ID for an access token.
func getUserID(accessToken string, client http.Client) (id string, err error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo?access_token=" + accessToken)
	if err != nil {
		return
	}
	var info UserInfo
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return
	}
	id = info.ID
	if id == "" {
		err = errors.New("Invalid auth")
	}
	return
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
	return "", int64(0), errors.New("Invalid path")
}

// datastoreAPI dispatches requests to the relevant API method and arranges certain common state
func datastoreAPI(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.Header().Add("Access-Control-Allow-Origin", "*")

	r.ParseForm()
	client := urlfetch.Client(c)

	// Get the access_token from the request and turn it into a user ID with which we will namespace Kinds in the datastore.
	accessToken := r.Form.Get("access_token")
	if accessToken == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := getUserID(accessToken, *client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kind, id, err := getKindAndID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dsKind := fmt.Sprintf("%s--%s", userID, kind)

	if id == int64(0) {
		switch r.Method {
		case "POST":
			insert(w, dsKind, r.Body, c)
			return
		case "GET":
			uq := userQuery(r)
			list(w, dsKind, uq, c)
			return
		}
	} else {
		switch r.Method {
		case "GET":
			get(w, dsKind, id, c)
			return
		case "DELETE":
			delete(w, dsKind, id, c)
			return
		case "POST":
			// This is strictly "replace all properties/values", not "add new properties, update existing"
			update(w, dsKind, id, r.Body, c)
			return
		}
	}
	http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
}

func userQuery(r *http.Request) UserQuery {
	uq := UserQuery{
		StartCursor: r.FormValue("start"),
		EndCursor: r.FormValue("end"),
	}
	// TODO: MustParse for limit/offset, else panic
	uq.Limit, _ = strconv.Atoi(r.FormValue("limit"))
	uq.Offset, _ = strconv.Atoi(r.FormValue("offset"))

	// TODO: Support ?where=foo<bar queries (which may or may not be annoying to scope for users...)
	_ = r.FormValue("where")
	return uq
}

func delete(w http.ResponseWriter, kind string, id int64, c appengine.Context) {
	k := datastore.NewKey(c, kind, "", id, nil)
	if err := datastore.Delete(c, k); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

func get(w http.ResponseWriter, kind string, id int64, c appengine.Context) {
	k := datastore.NewKey(c, kind, "", id, nil)
	var plist datastore.PropertyList
	if err := datastore.Get(c, k, &plist); err != nil {
		if err == datastore.ErrNoSuchEntity {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	m := plistToMap(plist, k)
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

func insert(w http.ResponseWriter, kind string, r io.Reader, c appengine.Context) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[createdKey] = time.Now()

	plist := mapToPlist(m)

	k := datastore.NewIncompleteKey(c, kind, nil)
	k, err := datastore.Put(c, k, &plist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = k.IntID()
	json.NewEncoder(w).Encode(m)
}

// plistToMap transforms a PropertyList such as you would get from the datastore into a map[string]interface{} suitable for JSON-encoding.
func plistToMap(plist datastore.PropertyList, k *datastore.Key) map[string]interface{} {
	m := make(map[string]interface{})
	for _, p := range plist {
		if _, exists := m[p.Name]; exists {
			if _, isArr := m[p.Name].([]interface{}); isArr {
				m[p.Name] = append(m[p.Name].([]interface{}), p.Value)
			} else {
				m[p.Name] = []interface{}{m[p.Name], p.Value}
			}
		} else {
			m[p.Name] = p.Value
		}
	}
	m[idKey] = k.IntID()
	return m
}

// mapToPlist transforms a map[string]interface{} such as you would get from decoding JSON into a PropertyList to store in the datastore.
func mapToPlist(m map[string]interface{}) datastore.PropertyList {
	plist := make(datastore.PropertyList, 0, len(m))
	for k, v := range m {
		if _, mult := v.([]interface{}); mult {
			for _, mv := range v.([]interface{}) {
				plist = append(plist, datastore.Property{
					Name:     k,
					Value:    mv,
					Multiple: true,
				})
			}
		} else {
			plist = append(plist, datastore.Property{
				Name:  k,
				Value: v,
			})
		}
	}
	return plist
}

func list(w http.ResponseWriter, kind string, uq UserQuery, c appengine.Context) {
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
		var plist datastore.PropertyList
		k, err := t.Next(&plist)
		if err == datastore.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m := plistToMap(plist, k)
		items = append(items, m)
		if crs, err = t.Cursor(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	r := map[string]interface{}{
		"items":          items,
		"nextStartToken": crs.String(),
	}
	json.NewEncoder(w).Encode(r)
}

func update(w http.ResponseWriter, kind string, id int64, r io.Reader, c appengine.Context) {
	var m map[string]interface{}
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[updatedKey] = time.Now()

	plist := mapToPlist(m)

	k := datastore.NewKey(c, kind, "", id, nil)
	if _, err := datastore.Put(c, k, &plist); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m[idKey] = id
	json.NewEncoder(w).Encode(m)
}
