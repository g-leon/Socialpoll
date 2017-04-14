package main

import (
	"context"
	"net/http"
	"gopkg.in/mgo.v2"
	"flag"
	"log"
)

// Server is the API server.
// Server makes sure handlers will not make
// database management mistakes.
type Server struct {
	db *mgo.Session
}

// contextKey helps to create uniform keys for
// context.Context, where keys are of type interface{}
type contextKey struct {
	name string
}

var contextKeyAPIKey = &contextKey{"api-key"}

func main() {
	var (
		addr = flag.String("addr", ":8080", "endpoint address")
		mongo = flag.String("mongo", "localhost", "mongodb address")
	)

	log.Println("Dialing mongo", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("failed to connect to mongo:", err)
	}
	defer db.Close()

	s := &Server{
		db: db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
	log.Println("Starting web service on", *addr)
	// TODO read
	http.ListenAndServe(":8080", mux)
	log.Println("Stopping...")
}

// APIKey is a helper function that, given a context,
// extracts a key.
func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

// withAPIKey is a wrapper of a HandlerFunc that helps with
// asking clients to provide an API key which facilitates the
// implementation of user authentication and authorisation.
func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}

func isValidAPIKey(key string) bool {
	// TODO check given key against a value read from a config file or database
	return key == "abc123"
}

// withCORS let's one to circumnavigate the same-origin policy,
// by being able to serve websites hosted on other domains as well.
func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(w, r)
	}
}