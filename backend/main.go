package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/dgrijalva/jwt-go"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net"
	"github.com/eawsy/aws-lambda-go-net/service/lambda/runtime/net/apigatewayproxy"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var (
	spotifyAuth spotify.Authenticator
	Handle      apigatewayproxy.Handler
	sess        *session.Session
	ddb         *dynamodb.DynamoDB

	jwtSecret []byte

	baseURL string
)

func init() {
	jwtSecret = []byte("totessecret")
	baseURL = os.Getenv("BASE_URL")

	var ddbEndpoint *string
	if baseURL == "http://localhost:3000" {
		ddbEndpoint = aws.String("http://dynamodb:8000")
	}
	sess = session.Must(session.NewSession(&aws.Config{
		Endpoint: ddbEndpoint,
	}))
	ddb = dynamodb.New(sess)

	spotifyAuth = spotify.NewAuthenticator(baseURL+"/spotifycallback", spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadEmail)

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

		log.Println("Authorized a new user")

		client := spotifyAuth.NewClient(token)
		cu, err := client.CurrentUser()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to retrieve current user from Spotify")
			http.Error(w, fmt.Sprintf("Auth error: %s", err), http.StatusInternalServerError)
			return
		}

		if err := saveSpotifyToken(cu.ID, token); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to save Spotify token")
			http.Error(w, fmt.Sprintf("Auth error: %s", err), http.StatusInternalServerError)
			return
		}

		// Generate a JWT token for cu.ID, redirect back to homepage with JWT stuffed into a cookie.
		userTok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
			Subject:   cu.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
			NotBefore: time.Now().Unix(),
		})

		tokCookie, err := userTok.SignedString(jwtSecret)
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to generate JWT after Spotify login")
			http.Error(w, fmt.Sprintf("Auth error: %s", err), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Path:  "/",
			Name:  "token",
			Value: tokCookie,
		})
		http.Redirect(w, r, baseURL, http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			panic(err)
		}

		var claims jwt.StandardClaims
		_, err = jwt.ParseWithClaims(cookie.Value, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return jwtSecret, nil
		})
		if err != nil {
			panic(err)
		}

		spotifyID := claims.Subject
		spotifyTok, err := getSpotifyToken(spotifyID)
		if err != nil {
			panic(err)
		}

		fmt.Fprint(w, "[")
		json.NewEncoder(w).Encode(spotifyTok)
		fmt.Fprint(w, ",")

		cli := spotifyAuth.NewClient(spotifyTok)
		cu, err := cli.CurrentUser()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(cu)
		fmt.Fprint(w, ",")

		json.NewEncoder(w).Encode(spotifyTok)

		fmt.Fprint(w, "]")
	})

	go http.Serve(ln, http.DefaultServeMux)
}

func saveSpotifyToken(id string, tok *oauth2.Token) error {
	tokattr, err := dynamodbattribute.Marshal(tok)
	if err != nil {
		return err
	}

	_, err = ddb.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("users"),
		Item: map[string]*dynamodb.AttributeValue{
			"id":    &dynamodb.AttributeValue{S: aws.String(id)},
			"token": tokattr,
		},
	})
	return err
}

func getSpotifyToken(id string) (tok *oauth2.Token, err error) {
	var res *dynamodb.GetItemOutput
	res, err = ddb.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{S: aws.String(id)},
		},
	})

	err = dynamodbattribute.Unmarshal(res.Item["token"], &tok)
	return
}
