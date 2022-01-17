package metastore

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func buildAwsCfg(uri string) (*url.URL, aws.Config, error) {
	var awsCfg aws.Config

	u, err := url.Parse(uri)
	if err != nil {
		return nil, awsCfg, err
	}

	ctx := context.Background()

	awsCfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, awsCfg, err
	}

	if u.Query().Has("region") {
		awsCfg.Region = u.Query().Get("region")
	}

	return u, awsCfg, nil
}
