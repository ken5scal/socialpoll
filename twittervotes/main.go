package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"net/url"
	"strings"
	"net/http"
	"encoding/json"
	"time"
	"github.com/bitly/go-nsq"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"bufio"
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
		options = append(options, p.Options...)
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

// Read(Seach) from Twitter
// votes channel is send-only channel
func readFromTwitter(votes chan <- string) {

	// load polling options
	options, err := loadOptions()
	if err != nil {
		log.Println("Failed reading polling options", err)
		return
	}

	hashtags := make([]string, len(options))
	for i := range options {
		hashtags[i] = "#" + strings.ToLower(options[i])
	}

	// Prase URL
	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("Failed in parsing URL", err)
		return
	}

	// Generate query object and Place into Request object
	query := url.Values{"track": {strings.Join(hashtags, ",")}}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("creating filter request failed:", err)
	}

	//query.Set("track", strings.Join(options, ","))
	//req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	//if err != nil {
	//	log.Println("Failed generating search request", err)
	//	return
	//}

	// Send reuest
	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("Failed requst, err")
		return
	}

	if resp.StatusCode != http.StatusOK {
		// this is a nice way to see what the error actually is:
		s := bufio.NewScanner(resp.Body)
		s.Scan()
		log.Println(s.Text())
		log.Println(hashtags)
		log.Println("StatusCode =", resp.StatusCode)
		return
	}

	// Decoder Response
	reader = resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var t tweet
		if err := decoder.Decode(&t); err == nil {
			for _, option := range options {
				if strings.Contains(
					strings.ToLower(t.Text),
					strings.ToLower(option),
				) {
					log.Println("vote:", option)
					votes <- option
				}
			}
		} else {
			break
		}
	}
}

// Signal Channel
// stopchan is receive-only channel
func startTwitterStream(stopchan <- chan struct{}, votes chan <- string) <- chan struct{} {
	// buffer size 1: if someone writes to channel, then channel will be blocked from writing until somone reads signal
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()

		for {
			// Wail for messages into the channel(stopchan)
			select {
			case <-stopchan:
			// Kill gorountine
				log.Println("Finishinq a query to Twitter")
				return
			default:
			// Notify goroutine killed
				log.Println("Starting a query to Twitter")
				readFromTwitter(votes) // receives polling options from db and coordinate twitter serching
				log.Println(" (Waiting)")
				time.Sleep(10 * time.Second) // Wait and Reconnect
			}

		}
	}()
	return stoppedchan
}


// Send search result tweet with selection to NSQ
func publishVotes(votes <-chan string) <- chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		defer func() {
			stopchan <- struct{}{}
		}()

		// periodically check votes channel
		// if channel is closed, loop will be terminated
		for vote := range votes {
			log.Println("Periodical publishing...")
			pub.Publish("votes", []byte(vote)) // Publish polling result
		}

		log.Println("Publisher: Stopping...")
		pub.Stop()
		log.Println("Publisher: Stopped")
	}()
	return stopchan
}

func main() {
	// Periodically fetch inspecing indexes from MongoDB and renew connections to Twitter
	// Multiple goroutine can call this method
	var stoplock sync.Mutex
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan //attempts reading from channel
		// following line will be executed only if the signal is eithe sigint or sigterm
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM) // sending a signal to signalChan when someone attempts to stop the program

	if err := dialdb(); err != nil {
		log.Fatalln("Failed dialing to MongoDB: ", err)
	}
	defer closedb()

	// Restart readFromTwitter evry minute in order to make sure up-to-date polling options from db
	// by calling startTwitterStream
	votes := make(chan string) // channel for voting result
	publisherStopChan := publishVotes(votes)
	twitterStopChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				break
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStopChan
	close(votes)
	<-publisherStopChan
}