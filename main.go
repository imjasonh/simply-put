package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

const (
	idKey        = "_id"
	createdKey   = "_created"
	updatedKey   = "_updated"
	defaultLimit = 10
)

var (
	port = flag.Int("port", 8080, "port to run on")
	db   = flag.String("db", "bolt.db", "bolt db file")

	invalidPath = errors.New("invalid path")
	nowFunc     = time.Now
)

func main() {
	flag.Parse()
	db, err := bolt.Open(*db, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s := &server{db}
	log.Println("server start")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), s))
}

type server struct {
	db *bolt.DB
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// TODO: user ID namespacing / auth

	kind, id, err := getKindAndID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var resp map[string]interface{}
	errCode := http.StatusOK
	if id == int64(0) {
		switch r.Method {
		case "POST":
			resp, errCode = s.insert(kind, r.Body)
			r.Body.Close()
		case "GET":
			uq, err := newUserQuery(r)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			resp, errCode = s.list(kind, *uq)
		default:
			http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			return
		}
	} else {
		switch r.Method {
		case "GET":
			resp, errCode = s.get(kind, id)
		case "DELETE":
			errCode = s.delete2(kind, id)
		case "POST":
			// This is strictly "replace all properties/values", not "add new properties, update existing"
			resp, errCode = s.update(kind, id, r.Body)
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
	if resp != nil && len(resp) != 0 {
		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Add("Content-Type", "application/json")
}

// getKindAndID parses the kind and ID from a request path.
func getKindAndID(path string) (string, int64, error) {
	if !strings.HasPrefix(path, "/") || path == "/" {
		return "", int64(0), invalidPath
	}
	parts := strings.Split(path[1:], "/")
	if len(parts) > 2 {
		return "", int64(0), invalidPath
	} else if len(parts) == 1 {
		return parts[0], int64(0), nil
	} else if len(parts) == 2 {
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", int64(0), err
		}
		return parts[0], id, nil
	}
	return "", int64(0), invalidPath
}

type filter struct {
	Key, Value string
}
type userQuery struct {
	Limit                        int
	StartCursor, EndCursor, Sort string
	Filters                      []filter
}

func newUserQuery(r *http.Request) (*userQuery, error) {
	uq := userQuery{
		StartCursor: r.FormValue("start"),
		EndCursor:   r.FormValue("end"),
		Sort:        r.FormValue("sort"),
	}
	if r.FormValue("limit") != "" {
		lim, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			return nil, err
		}
		uq.Limit = lim
	}

	for _, f := range map[string][]string(r.Form)["where"] {
		parts := strings.Split(f, "=")
		if len(parts) != 2 {
			return nil, errors.New("invalid where: " + f)
		}
		uq.Filters = append(uq.Filters, filter{Key: parts[0], Value: parts[1]})
	}
	return &uq, nil
}

func (s *server) delete2(kind string, id int64) int {
	// TODO: implement
	return http.StatusOK
}

func (s *server) get(kind string, id int64) (map[string]interface{}, int) {
	// TODO: implement
	return nil, http.StatusOK
}

func (s *server) insert(kind string, r io.Reader) (map[string]interface{}, int) {
	m, err := fromJSON(r)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	m[createdKey] = nowFunc().Unix()

	// TODO: implement
	return m, http.StatusOK
}

func (s *server) list(kind string, uq userQuery) (map[string]interface{}, int) {
	// TODO: implement
	items := make([]map[string]interface{}, 0)
	r := map[string]interface{}{
		"items": items,
	}
	return r, http.StatusOK
}

func (s *server) update(kind string, id int64, r io.Reader) (map[string]interface{}, int) {
	m, err := fromJSON(r)
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	delete(m, createdKey) // Ignore any _created value the user provides
	delete(m, idKey)      // Ignore any _id value the user provides
	m[updatedKey] = nowFunc().Unix()

	// TODO: implement
	return m, http.StatusOK
}

func fromJSON(r io.Reader) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.NewDecoder(r).Decode(&m)
	if err != nil {
		fmt.Errorf("decoding json: %v", err)
	}
	return m, err
}
