package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Grab the oauth code sent back from our IAM server redirect
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Oops! No code found in the URL. Try again.")
			return
		}

		// Just dump the raw code on the screen so we can yoink it for Postman
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Boom! Here is your OAuth Code: %s\n\nCopy that and head over to Postman to swap it for a TokenPair!", code)
	})

	// Spin up a dummy server on port 9999 to catch the ball
	fmt.Println("🚀 Quick test app is up and running on: http://127.0.0.1:9999")
	if err := http.ListenAndServe("127.0.0.1:9999", nil); err != nil {
		panic(err)
	}
}
