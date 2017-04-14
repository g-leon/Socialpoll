package main

import (
	"encoding/json"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/joeshaw/envdecode"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	conn          net.Conn
	reader        io.ReadCloser
	authClient    *oauth.Client
	creds         *oauth.Credentials
	authSetupOnce sync.Once
	httpClient    *http.Client
)

type tweet struct {
	Text string
}

func startTwitterStream(stopChan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopChan:
				log.Println("stopping Twitter...")
				return
			default:
				log.Println("Querying Twitter...")
				readFromTwitter(votes)
				log.Println("    (waiting)")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stoppedchan
}

// readFromTwitter reloads the options from the database
// each time it is called so the the program is updated without
// having to restart it.
func readFromTwitter(votes chan<- string) {
	options, err := loadOptions()
	if err != nil {
		log.Println("failed to load options:", err)
		return
	}
	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("creating filter request failed:", err)
		return
	}
	query := make(url.Values)
	query.Set("track", strings.Join(options, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("creating filter request failed:", err)
		return
	}
	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("making request failed:", err)
		return
	}
	reader := resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var t tweet
		if err := decoder.Decode(&t); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(
				strings.ToLower(t.Text),
				strings.ToLower(option),
			) {
				log.Println("vote:", option)
				votes <- option
			}
		}
	}
}

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
		ConsumerKey    string `env:"SP_TWITTER_KEY,required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET,required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN,required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET,required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

// makeRequest makes sure the initialisation code runs only once.
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
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))

	return httpClient.Do(req)
}
