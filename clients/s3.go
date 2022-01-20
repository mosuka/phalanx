package clients

import (
	"net/url"
	"os"

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

	// Create AWS config.
	cfg, err := NewAwsConfigWithUri(uri)
	if err != nil {
		return nil, err
	}

	usePathStyle := false
	usePathStyleStr := os.Getenv("AWS_USE_PATH_STYLE")
	if str := u.Query().Get("use_path_style"); str != "" {
		usePathStyleStr = str
	}
	if usePathStyleStr == "true" {
		usePathStyle = true
	}

	return s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = usePathStyle
	}), nil
}
