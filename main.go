package main

import (
	"errors"
	"flag"
	"fmt"
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

func (s *server) delete2(kind, id string) int {
	tx, err := s.db.Begin(true)
	if err != nil {
		log.Printf("begin tx: %v", err)
		return http.StatusInternalServerError
	}
	b := tx.Bucket([]byte(kind))
	if b == nil {
		return http.StatusNotFound
	}
	if err := b.Delete([]byte(id)); err != nil {
		log.Printf("delete: %v", err)
		return http.StatusInternalServerError
	}
	if err := tx.Commit(); err != nil {
		log.Printf("commit delete: %v", err)
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func (s *server) get(kind, id string) ([]byte, int) {
	tx, err := s.db.Begin(false)
	if err != nil {
		log.Printf("begin tx: %v", err)
		return nil, http.StatusInternalServerError
	}
	b := tx.Bucket([]byte(kind))
	if b == nil {
		return nil, http.StatusNotFound
	}
	v := b.Get([]byte(id))
	if v == nil {
		return nil, http.StatusNotFound
	}
	if err := tx.Commit(); err != nil {
		log.Printf("commit get: %v", err)
		return nil, http.StatusInternalServerError
	}
	return v, http.StatusOK
}

func (s *server) insert(kind string, r io.Reader) ([]byte, int) {
	// TODO: add _created ?
	tx, err := s.db.Begin(true)
	if err != nil {
		log.Printf("begin tx: %v", err)
		return nil, http.StatusInternalServerError
	}
	b, err := tx.CreateBucketIfNotExists([]byte(kind))
	if err != nil {
		log.Printf("create bucket: %v", err)
		return nil, http.StatusInternalServerError
	}
	var k []byte
	for {
		u, err := uuid.NewV5(uuid.NamespaceURL, []byte("imjasonh.com"))
		if err != nil {
			log.Printf("uuid: %v", err)
			return nil, http.StatusInternalServerError
		}
		k = u[:]
		if conflict := b.Get(k); conflict == nil {
			break
		}
	}
	all, err := ioutil.ReadAll(r)
	if err != nil {
		log.Printf("readall: %v", err)
		return nil, http.StatusInternalServerError
	}
	// TODO: add _id to the JSON
	if err := b.Put(k, all); err != nil {
		log.Println("put: %v", err)
		return nil, http.StatusInternalServerError
	}
	if err := tx.Commit(); err != nil {
		log.Printf("commit put: %v", err)
		return nil, http.StatusInternalServerError
	}
	return all, http.StatusOK
}

func (s *server) list(kind string, uq userQuery) ([]byte, int) {
	tx, err := s.db.Begin(false)
	if err != nil {
		log.Printf("begin tx: %v", err)
		return nil, http.StatusInternalServerError
	}
	b := tx.Bucket([]byte(kind))
	if b == nil {
		return nil, http.StatusNotFound
	}
	_ = b.Cursor()
	// TODO: implement
	if err := tx.Commit(); err != nil {
		log.Printf("commit delete: %v", err)
		return nil, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (s *server) update(kind, id string, r io.Reader) ([]byte, int) {
	tx, err := s.db.Begin(true)
	if err != nil {
		log.Printf("begin tx: %v", err)
		return nil, http.StatusInternalServerError
	}
	b := tx.Bucket([]byte(kind))
	if b == nil {
		return nil, http.StatusNotFound
	}
	k := []byte(id)
	v := b.Get(k)
	if v == nil {
		return nil, http.StatusNotFound
	}
	all, err := ioutil.ReadAll(r)
	if err != nil {
		log.Printf("readall: %v", err)
		return nil, http.StatusInternalServerError
	}
	if err := b.Put(k, all); err != nil {
		log.Println("put: %v", err)
		return nil, http.StatusInternalServerError
	}
	if err := tx.Commit(); err != nil {
		log.Printf("commit update: %v", err)
		return nil, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
