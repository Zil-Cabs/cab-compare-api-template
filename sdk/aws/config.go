package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Ideally should be used only once throughout the app.
// and then should be cached.
func CreateAWSConfig(region string) aws.Config {
	newAwsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("could not get awsconfig")
	}
	newAwsCfg.Region = region
	return newAwsCfg
}
