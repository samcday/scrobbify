package main

import (
	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/eawsy/aws-lambda-go-event/service/lambda/runtime/event/apigatewayproxyevt"
)

type response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func Handle(evt *apigatewayproxyevt.Event, ctx *runtime.Context) (response, error) {
	return response{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/html"}, Body: "<h1>Hello, world!</h1>"}, nil
}
