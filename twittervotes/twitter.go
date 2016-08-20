package main

import (
	"net"
	"time"
	"io"
	"log"
)

var conn net.Conn
func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5 * time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

// Not Only closing connection, but also closing reader which reads response
var reader io.ReadCloser
func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

// Setup Oauth Object for authenticating request
var (
	authClient *oauth.Client
	creds *oauth.Credentials
)
func setupTwitterAuth() {
	var ts struct {
		ConsumerKey string `env:"SP_TWITTER_KEY=,required"`
		ConsumerSecret string `env:"SP_TWITTER_KEY=,required"`
		AccessToken string `env:"SP_TWITTER_ACCESSTOKEN=,required"`
		AccessSecret string `env:"SP_TWITTER_ACCESSECRET==,required"`
	}

	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}

	creds = &oauth.Credentials {
		Token: ts.AccessToken,
		Secret: ts.AccessSecret,
	}

	authClient = &oauth.Client {
		Token: ts.ConsumerKey,
		Secret: ts.ConsumerSecret,
	}
}