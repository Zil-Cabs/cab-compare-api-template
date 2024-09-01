package websocket

import (
	awsconfig "TODO/sdk/aws"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

type OnSocketRequestFn func(context.Context, events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error)

func GetHandlerFn(
	onConnect OnSocketRequestFn,
	onDisconnect OnSocketRequestFn,
	onMessage OnSocketRequestFn,
) OnSocketRequestFn {

	return func(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		switch request.RequestContext.RouteKey {
		case "$connect":
			// TODO add, auth etc prebuilt fns
			return onConnect(ctx, request)
		case "$disconnect":
			// TODO add cleanup prebuilt fns
			return onDisconnect(ctx, request)
		case "$default":
			return onMessage(ctx, request)
		default:
			bts, _ := json.Marshal(request)
			log.Println("error", "unhandled route key", request.RequestContext.RouteKey, string(bts))
			return events.APIGatewayProxyResponse{}, errors.New("unhandled route key")
		}
	}
}

func GetApigwMgmtClient(websocketApiId string) *apigatewaymanagementapi.Client {
	return apigatewaymanagementapi.NewFromConfig(
		awsconfig.CreateAWSConfig(os.Getenv("PROVIDER_REGION")),
		func(o *apigatewaymanagementapi.Options) {
			url := fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", websocketApiId, os.Getenv("PROVIDER_REGION"), os.Getenv("PROVIDER_STAGE"))
			o.BaseEndpoint = &url
		})
}
