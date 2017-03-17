package main

import (
	"io"
	"net"
	"time"
	"log"
	"github.com/joeshaw/envdecode"
	"github.com/garyburd/go-oauth/oauth"
	"sync"
	"net/http"
	"net/url"
	"strconv"
)

var (
	conn       net.Conn
	reader     io.ReadCloser
	authClient *oauth.Client
	creds      *oauth.Credentials
	authSetupOnce sync.Once
	httpClient *http.Client
)

// dial ensures that the connection is closed and
// then opens a new connection, keeping the conn
// variable updated with the new connection.
// If a connection dies or is closed, we can
// redial without worrying about zombie connections.
func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	conn = netc

	return netc, nil
}

// closeConn can be called at any time in order to break
// the ongoing connection with Twitter and tidy things up.
// If the program is called with Ctrl+C then we can call
// this function just before exiting.
func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

// setupTwitterAuth reads the environment variables and
// sets up the OAuth object needed in order to
// authenticate requests.
func setupTwitterAuth() {
	var ts struct {
		ConsumerKey    string `env:"SP_TWITTER_KEY, required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET, required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN, required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET, required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	creds = &oauth.Credentials{
		Token: ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token: ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

// makeRequest makes sure the initialisation code runs only once.
func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func () {
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
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))

	return httpClient.Do(req)
}