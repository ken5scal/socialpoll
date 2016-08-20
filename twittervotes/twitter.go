package main

import (
	"net"
	"time"
	"io"
	"log"
	"github.com/joeshaw/envdecode"
	"github.com/garyburd/go-oauth/oauth"
	"sync"
	"net/http"
	"net/url"
	"strconv"
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
		// back-quoted part is called tag. By using reflection API, one can access to the tag
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
		Credentials: oauth.Credentials{
			Token: ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

// Following authentication, this method creates request and receives response
var (
	authSetupOnce sync.Once // Singleton?
	httpClient *http.Client
)

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})

	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization",
		authClient.AuthorizationHeader(creds, "POST", req.URL, params))
	return httpClient.Do(req)
}