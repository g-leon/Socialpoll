package main

import "context"

// contextKey helps to create uniform keys for
// context.Context, where keys are of type interface{}
type contextKey struct {
	name string
}

var contextKeyAPIKey = &contextKey{"api-key"}

// APIKey is a helper function that, given a context,
// extracts a key.
func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

func main() {

}
