// Lambda entrypoint for LocalStack / AWS (provided.al2 runtime).
// Build: GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./lambda
package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body := map[string]any{
		"ok":     true,
		"path":   req.Path,
		"method": req.HTTPMethod,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: `{"error":"marshal"}`}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(b),
	}, nil
}

func main() {
	lambda.Start(handler)
}
