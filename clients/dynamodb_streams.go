package clients

import (
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/mosuka/phalanx/errors"
)

func NewDynamoDBStreamsClientWithUri(uri string) (*dynamodbstreams.Client, error) {
	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "dynamodb" {
		return nil, errors.ErrInvalidUri
	}

	// Create AWS config.
	cfg, err := NewAwsConfigWithUri(uri)
	if err != nil {
		return nil, err
	}

	return dynamodbstreams.NewFromConfig(cfg), nil
}
