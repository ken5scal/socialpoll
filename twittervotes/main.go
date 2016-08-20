package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"net/url"
	"strings"
	"net/http"
	"encoding/json"
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

// Receive send-only channel(votes) and notify polling events happened on Twitter
type tweet struct {
	Text string
}

func readFromTwitter(votes chan <- string) {
	options, err := loadOptions()
	if err != nil {
		log.Println("Failed reading polling options", err)
		return
	}

	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("Failed in parsing URL", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(),
		strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("Failed generating search request", err)
		return
	}
	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("Failed requst, err")
		return
	}

	reader = resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var tweet tweet
		if err := decoder.Decode(&tweet); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(strings.ToLower(tweet.Text), strings.ToLower(option)) {
				log.Println("Poll:", option)
				votes <- option
			}
		}
	}

}


// Send search result tweet with selection to NSQ

// Periodically fetch inspecing indexes from MongoDB and renew connections to Twitter

// When Ctrl + C, then finish program

