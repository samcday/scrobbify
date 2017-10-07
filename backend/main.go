package main

import (
	"fmt"
	"net/http"

	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	"github.com/zmb3/spotify"
)

// var spotifyAccountBaseUrl = "https://accounts.spotify.com"

var spotifyAuth spotify.Authenticator

// Handle is the exported handler called by AWS Lambda.
var Handle apigatewayproxy.Handler

func init() {
	spotifyAuth = spotify.NewAuthenticator("http://localhost:3000/spotifycallback", spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadEmail)

	ln := net.Listen()
	Handle = apigatewayproxy.New(ln, nil).Handle

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, spotifyAuth.AuthURL("bacon"), http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/spotifycallback", func(w http.ResponseWriter, r *http.Request) {
		token, err := spotifyAuth.Token("bacon", r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Auth error: %s", err), http.StatusNotFound)
			return
		}

		client := spotifyAuth.NewClient(token)

		cp, err := client.PlayerCurrentlyPlaying()
		fmt.Fprintf(w, "Yay! %s", cp.Item.Name)
	})

	go http.Serve(ln, http.DefaultServeMux)
}
