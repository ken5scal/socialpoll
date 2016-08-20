package main

import (
	"gopkg.in/mgo.v2"
	"log"
)

/*
- Read Mongodb to fetch polling options for twitter search
 */
var db *mgo.Session // Global object

// Connecting to DB
func dialdb() error {
	var err error
	log.Println("Dialing MongoDB: localhost")
	db, err = mgo.Dial("localhost")
	return err
}

// Closing DB connection
func closedb() {
	db.Close()
	log.Println("Closed db connection")
}

// Load polling options
type poll struct {
	Options []string
}
func loadOptions() ([]string, error) {
	var options []string
	var p poll

	// ballots -> DB
	// polls -> Collection
	// Find(nil -> search without filtering
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	for iter.Next(&p) {
		options = append(options, p.Options)
	}
	iter.Close()
	return options, iter.Err()
}


/*
- Read from Twitter
	- Use Twitter streaming api to manage session
	- search for tweet responding to the polls
 */



// Send search result tweet with selection to NSQ

// Periodically fetch inspecing indexes from MongoDB and renew connections to Twitter

// When Ctrl + C, then finish program

