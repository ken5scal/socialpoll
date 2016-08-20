package main

import (
	"gopkg.in/mgo.v2"
	"log"
)


// read all polling result and retrieve all polls
// from $options array in each document using mgo via MongoDB
var db *mgo.Session

func dialdb() error {
	var err error
	log.Println("Dialing MongoDB: localhost")
	db, err = mgo.Dial("localhost")
	return err
}

func closedb() {
	db.Close()
	log.Println("Closed db connection")
}


// Use Twitter streaming api to manage session and search for tweet responding to the polls

// Send search result tweet with selection to NSQ

// Periodically fetch inspecing indexes from MongoDB and renew connections to Twitter

// When Ctrl + C, then finish program

