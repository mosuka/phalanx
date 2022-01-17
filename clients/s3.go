package clients

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mosuka/phalanx/errors"
)

func NewS3ClientWithUri(uri string) (*s3.Client, error) {
	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "s3" {
		return nil, errors.ErrInvalidUri
	}

	ctx := context.Background()

	var cfg aws.Config

	profile := os.Getenv("AWS_PROFILE")
	if str := u.Query().Get("profile"); str != "" {
		profile = str
	}
	if profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
	}

	useCredentials := false
	accessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	if str := u.Query().Get("access_key_id"); str != "" {
		accessKeyId = str
	}
	if accessKeyId != "" {
		useCredentials = true
	}

	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if str := u.Query().Get("secret_access_key"); str != "" {
		secretAccessKey = str
	}
	if secretAccessKey != "" {
		useCredentials = true
	}

	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	if str := u.Query().Get("session_token"); str != "" {
		sessionToken = str
	}
	if sessionToken != "" {
		useCredentials = true
	}

	if useCredentials {
		cfg.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, sessionToken))
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if str := u.Query().Get("region"); str != "" {
		region = str
	}
	if region != "" {
		cfg.Region = region
	}

	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if str := u.Query().Get("endpoint_url"); str != "" {
		endpoint = str
	}
	if endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           endpoint,
				SigningRegion: cfg.Region,
			}, nil
		})
		cfg.EndpointResolverWithOptions = customResolver
	}

	usePathStyle := false
	usePathStyleStr := os.Getenv("AWS_USE_PATH_STYLE")
	if str := u.Query().Get("use_path_style"); str != "" {
		usePathStyleStr = str
	}
	if usePathStyleStr == "true" {
		fmt.Println("AWS_USE_PATH_STYLE is set to true. Using path style.")
		usePathStyle = true
	}

	return s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = usePathStyle
	}), nil
}
