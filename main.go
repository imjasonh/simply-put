package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

var (
	port = flag.Int("port", 8080, "port to run on")
	db   = flag.String("db", "bolt.db", "bolt db file")
)

func main() {
	flag.Parse()
	db, err := bolt.Open(*db, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	s := &Server{db}
	log.Println("server start")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), s))
}
