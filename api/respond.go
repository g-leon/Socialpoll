package main

import (
	"net/http"
	"encoding/json"
	"fmt"
)

// decodeBody abstracts away the message decoding part,
// such that one can easily change the way messages are encoded
// and decoded.
func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// encodeBody abstracts away the message encoding part,
// such that one can easily change the way messages are encoded
// and decoded.
func encodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

// respond makes it easy to write the status code and some data
// to the ResponseWriter object using the encodeBody helper.
func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, r, data)
	}
}

// respondErr is a helper that abstracts the error responding.
// It is an interface similar to the respond function,
// but the data written will be enveloped in an error object
// in order to make it clear that something went wrong.
func respondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	respond(w, r, status, map[string]interface{} {
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),
		},
	})
}


// respondHTTPErr is an HTTP-error-specific helper that
// will generate the correct message, using the http.StatusText
// function from the standard library.
func respondHTTPErr(w http.ResponseWriter, r *http.Request, status int) {
	respondErr(w, r, status, http.StatusText(status))
}