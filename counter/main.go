package main

import (
	"fmt"
	"flag"
	"os"
	"log"
	"gopkg.in/mgo.v2"
)

var fatalErr error
func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func main() {
	defer func() { // <- defer 1
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("Connecting db...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return // ensures to finish main (call defer funcs)
	}

	defer func() { // <- defer 2. called before defer 1. LIFO
		log.Println("Disconnecting db...")
		db.Close()
	}()

	pollData := db.DB("ballots").C("polls")
}
