package metastore

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// we must have a region for setting up an endpoint when testing so we use the standard default of north virginia
const defaultSigningRegion = "us-east-1"

func buildAwsCfg(uri string) (*url.URL, aws.Config, error) {
	var awsCfg aws.Config

	u, err := url.Parse(uri)
	if err != nil {
		return nil, awsCfg, err
	}

	ctx := context.Background()

	awsCfg, err = config.LoadDefaultConfig(ctx, buildOpts(u)...)
	if err != nil {
		return nil, awsCfg, err
	}

	if u.Query().Has("region") {
		awsCfg.Region = u.Query().Get("region")
	}

	return u, awsCfg, nil
}

func buildOpts(u *url.URL) []func(*config.LoadOptions) error {
	opts := make([]func(*config.LoadOptions) error, 0)

	if creds := getCredentials(u); creds != nil {
		opts = append(opts, creds)
	}

	if endpoint := getEndpoint(u); endpoint != nil {
		opts = append(opts, endpoint)
	}

	return opts
}

func getCredentials(u *url.URL) config.LoadOptionsFunc {
	switch {
	case u.Query().Has("profile"):
		return config.WithSharedConfigProfile("test-profile")
	case u.Query().Has("access_key_id"):
		return config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			u.Query().Get("access_key_id"),
			u.Query().Get("secret_access_key"),
			u.Query().Get("session_token"),
		))
	default:
		return nil
	}
}

func getEndpoint(u *url.URL) config.LoadOptionsFunc {
	if u.Query().Has("endpoint_url") {

		signingRegion := defaultSigningRegion

		if u.Query().Has("region") {
			signingRegion = u.Query().Get("region")
		}

		return config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           u.Query().Get("endpoint_url"),
						SigningRegion: signingRegion,
					}, nil
				},
			),
		)
	}

	return nil
}
