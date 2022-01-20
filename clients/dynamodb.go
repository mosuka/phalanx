package clients

import (
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mosuka/phalanx/errors"
)

func NewDynamoDBClientWithUri(uri string) (*dynamodb.Client, error) {
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

	return dynamodb.NewFromConfig(cfg), nil
}
