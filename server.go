package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/nu7hatch/gouuid"
)

const (
	idKey        = "_id"
	createdKey   = "_created"
	updatedKey   = "_updated"
	defaultLimit = 10
)

var (
	invalidPath = errors.New("invalid path")
	nowFunc     = time.Now
)

type Server struct {
	db *bolt.DB
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// TODO: user ID namespacing / auth

	kind, id, err := getKindAndID(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var b []byte
	errCode := http.StatusOK
	if id == "" {
		switch r.Method {
		case "POST":
			b, errCode = s.insert(kind, r.Body)
			r.Body.Close()
		case "GET":
			uq, err := newUserQuery(r)
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			b, errCode = s.list(kind, *uq)
		default:
			http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
			return
		}
	} else {
		switch r.Method {
		case "GET":
			b, errCode = s.get(kind, id)
		case "DELETE":
			errCode = s.delete2(kind, id)
		case "POST":
			// This is strictly "replace all properties/values", not "add new properties, update existing"
			b, errCode = s.update(kind, id, r.Body)
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
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

// getKindAndID parses the kind and ID from a request path.
func getKindAndID(path string) (string, string, error) {
	if !strings.HasPrefix(path, "/") || path == "/" {
		return "", "", invalidPath
	}
	parts := strings.Split(path[1:], "/")
	if len(parts) > 2 {
		return "", "", invalidPath
	} else if len(parts) == 1 {
		return parts[0], "", nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	return "", "", invalidPath
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
		Limit:       defaultLimit,
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

func (s *Server) delete2(kind, id string) int {
	code := http.StatusOK
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		if b == nil {
			code = http.StatusNotFound
			return nil
		}
		if err := b.Delete([]byte(id)); err != nil {
			log.Printf("delete: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return http.StatusInternalServerError
	}
	return code
}

func (s *Server) get(kind, id string) (out []byte, code int) {
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		if b == nil {
			code = http.StatusNotFound
			return nil
		}
		out = b.Get([]byte(id))
		if out == nil {
			code = http.StatusNotFound
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	return
}

func (s *Server) insert(kind string, r io.Reader) (out []byte, code int) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(kind))
		if err != nil {
			log.Printf("create bucket: %v", err)
			return err
		}
		var k []byte
		for {
			u, err := uuid.NewV5(uuid.NamespaceURL, []byte("imjasonh.com"))
			if err != nil {
				log.Printf("uuid: %v", err)
				return err
			}
			k = u[:]
			if conflict := b.Get(k); conflict == nil {
				break
			}
		}
		all, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf("readall: %v", err)
			return err
		}
		m, err := fromJSON(all)
		if err != nil {
			log.Printf("json: %v", err)
			return err
		}
		m[idKey] = k
		m[createdKey] = nowFunc().Unix()
		if err := b.Put(k, all); err != nil {
			log.Printf("put: %v", err)
			return err
		}
		out, err = toJSON(m)
		if err != nil {
			log.Printf("json: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	return
}

func (s *Server) list(kind string, uq userQuery) (out []byte, code int) {
	code = http.StatusOK
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		if b == nil {
			code = http.StatusNotFound
			return nil
		}
		_ = b.Cursor()
		// TODO: implement
		return nil
	})
	if err != nil {
		return nil, http.StatusNotFound
	}
	return
}

func (s *Server) update(kind, id string, r io.Reader) (out []byte, code int) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(kind))
		if b == nil {
			code = http.StatusNotFound
			return nil
		}
		k := []byte(id)
		v := b.Get(k)
		if v == nil {
			code = http.StatusNotFound
			return nil
		}
		all, err := ioutil.ReadAll(r)
		if err != nil {
			log.Printf("readall: %v", err)
			return err
		}
		m, err := fromJSON(all)
		if err != nil {
			log.Printf("json: %v", err)
			return err
		}
		m[updatedKey] = nowFunc().Unix()
		if err := b.Put(k, all); err != nil {
			log.Printf("put: %v", err)
			return err
		}
		out, err = toJSON(m)
		if err != nil {
			log.Printf("json: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, http.StatusInternalServerError
	}
	return
}

func fromJSON(b []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.NewDecoder(bytes.NewReader(b)).Decode(&m)
	return m, err
}

func toJSON(m map[string]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(m)
	return buf.Bytes(), err
}
