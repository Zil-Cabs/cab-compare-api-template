package main

import (
	sdklambda "TODO/sdk/lambda"
	"TODO/sdk/websocket"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

var ginLambda *ginadapter.GinLambdaV2
var apiClient *apigatewaymanagementapi.Client

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowOriginFunc = func(origin string) bool {
		return strings.HasSuffix(origin, ".whichone.in") ||
			strings.HasPrefix(origin, "http://localhost:")
	}

	router.Use(cors.New(config))
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "world"})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "helloworld"})
	})

	ginLambda = ginadapter.NewV2(router)

	lambda.Start(sdklambda.MultiEventTypeHandler(
		[]sdklambda.OwnEventHandler{
			sdklambda.GetOwnEventHandler(
				events.APIGatewayWebsocketProxyRequest{},
				wsHandler(),
			),
			sdklambda.GetOwnEventHandler(
				events.APIGatewayV2HTTPRequest{},
				httpHandler,
			),
		},
	))
}

func httpHandler(ctx context.Context, httpReq events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, httpReq)
}

func wsHandler() websocket.OnSocketRequestFn {
	return websocket.GetHandlerFn(
		func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
			js, _ := json.Marshal(req)
			log.Println("req", string(js))
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}, nil
		},
		func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
			js, _ := json.Marshal(req)
			log.Println("req", string(js))
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}, nil
		},
		func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
			js, _ := json.Marshal(req)
			log.Println("req", string(js))
			if apiClient == nil {
				apiClient = websocket.GetApigwMgmtClient(os.Getenv("WEBSOCKET_API_ID"))
			}

			data, err := json.Marshal(req)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusOK,
				}, err
			}

			_, err = apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: &req.RequestContext.ConnectionID,
				Data:         data,
			})

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			}, err
		},
	)
}
