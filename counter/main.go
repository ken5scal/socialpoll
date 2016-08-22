package main

import (
	"fmt"
	"flag"
	"os"
	"log"
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/bitly/go-nsq"
	"time"
	"gopkg.in/mgo.v2/bson"
	"os/signal"
	"syscall"
)

const updateDuration = 1 * time.Second

var fatalErr error

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func main() {
	defer func() {
		// <- defer 1
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

	defer func() {
		// <- defer 2. called before defer 1. LIFO
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
		if counts == nil {
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

	// Pushes result to db periodically
	log.Println("Waiting for ballots on NSQ...")
	var updater *time.Timer
	updater = time.AfterFunc(updateDuration, func() { //loop inside func
		countsLock.Lock()
		defer countsLock.Unlock()
		if len(counts) == 0 {
			log.Println("No new polls. Skipping db update")
		} else {
			log.Println("updating db...")
			log.Println(counts)
			ok := true
			for option, count := range counts {
				sel := bson.M{"options": bson.M{"$in": []string{option}}}
				up := bson.M{"$inc": bson.M{"results." + option: count}}
				if _, err := pollData.UpdateAll(sel, up); err != nil {
					log.Println("Failed updating...", err)
					ok = false
					continue
				}
				counts[option] = 0
			}
			if ok {
				log.Println("Updated")
				counts = nil
			}
		}
		updater.Reset(updateDuration)
	})

	// Handles Ctrl + C
	termChan := make(chan os.Signal, 1) // channel to catch ctl+c
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-termChan:
			updater.Stop()
			q.Stop()
		case <-q.StopChan:
			return //finish
		}
	}
}
