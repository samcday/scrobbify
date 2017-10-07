package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
)

// Handle is the exported handler called by AWS Lambda.
var Handle apigatewayproxy.Handler

func init() {
	ln := net.Listen()
	Handle = apigatewayproxy.New(ln, nil).Handle
	go http.Serve(ln, http.DefaultServeMux)

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		qs := url.Values{}
		qs.Add("response_type", "code")
		qs.Add("client_id", os.Getenv("SPOTIFY_CLIENT_ID"))
		qs.Add("scope", "user-read-currently-playing user-read-email")
		qs.Add("redirect_uri", "http://localhost:3000/spotifycallback")
		qs.Add("state", "bacon")

		http.Redirect(w, r, "https://accounts.spotify.com/authorize?"+qs.Encode(), http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/spotifycallback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != "bacon" {
			fmt.Fprint(w, "Bad state")
			return
		}

		json.NewEncoder(w).Encode(r.URL.Query())
	})
}
