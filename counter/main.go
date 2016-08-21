package main

import (
	"fmt"
	"flag"
	"os"
	"log"
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/bitly/go-nsq"
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

	var countsLock sync.Mutex // map and lock. avoid read/write conflicts by multiple goroutine to a map
	var counts map[string]int
	log.Println("Connecting NSQ...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return // ensures to finish main (call defer funcs)
	}

	// Handle events receiving a message from NSQ
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts ==nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))

	// Connecting to NSQ Service
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}
}
