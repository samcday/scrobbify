# scrobbify

Synchronize listening activity on Spotify to Last.fm, using the [Spotify Web API](https://developer.spotify.com/web-api/) and the [Last.fm API](https://www.last.fm/api).

Built as a React frontend app + AWS SAM backend, with the added awesomeness of [aws-lambda-go-shim](https://github.com/eawsy/aws-lambda-go-shim).

## Setup

Prereqs:

 * AWS CLI
 * Docker + Docker-Compose
 * Go

```
go get github.com/cespare/reflex github.com/mattn/goreman

docker-compose up -d
./setup.sh
```

## Local development

Bring up SAM Local API server, a process to recompile on backend changes, etc:

```
goreman start
```

Then hit http://localhost:3000
