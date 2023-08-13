package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambdaV2

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
	lambda.Start(Handler)
}

func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return ginLambda.ProxyWithContext(ctx, request)
}
