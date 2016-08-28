package main

import (
	"net/http"
	"gopkg.in/mgo.v2"
	"flag"
	"log"
	"time"
	"github.com/stretchr/graceful"
)

func main() {
	var (
		addr = flag.String("addr", ":8080", "Endpoint Address")
		mongo = flag.String("mongo", "localhost", "MongoDB Address")
	)
	flag.Parse()
	log.Println("Connecting to MongoDB", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("Failed connecting to MongoDB: ", err)
	}
	defer db.Close()
	mux := http.NewServeMux()
	// Handle's request with a path starting from /polls/
	/*
		1. withCORS returns Handlers inserting Headers
		2. withVars returns Handlers mapping data in memory assuring to be released
		3. withData returns Handlers which coppied db session
		4. withAPIKey checks key
		5. then handlePolls is called
		6. All data (db session, data in memory) will be cleaned
	 */
	mux.HandleFunc("/polls/", withCORS(withVars(withData(db, withAPIKey(handlePolls)))))
	log.Println("Starting Web server: ", *addr)
	graceful.Run(*addr, 1*time.Second,mux)
	log.Println("Stopping...")
}

// Wrap HandlerFunc
func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isValidAPIKey(r.URL.Query().Get("key")) {
			respondErr(w, r, http.StatusUnauthorized, "Invalid API Key")
			return
		}
		fn(w, r)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}

func withData(d *mgo.Session, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		thisDb := d.Copy()
		defer thisDb.Close() // Copu db session
		SetVar(r, "db", thisDb.DB("ballots"))
		f(w, r)
	}
}