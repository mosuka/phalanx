package clients

import (
	"net/url"
	"os"
	"strconv"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mosuka/phalanx/errors"
)

func NewMinioClientWithUri(uri string) (*minio.Client, error) {
	// Parse URI.
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "minio" {
		return nil, errors.ErrInvalidUri
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if str := u.Query().Get("endpoint"); str != "" {
		endpoint = str
	}

	user := os.Getenv("AWS_ACCESS_KEY_ID")
	if str := os.Getenv("MINIO_ACCESS_KEY"); str != "" {
		user = str
	}
	if str := u.Query().Get("access_key"); str != "" {
		user = str
	}

	password := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if str := os.Getenv("MINIO_SECRET_KEY"); str != "" {
		password = str
	}
	if str := u.Query().Get("secret_key"); str != "" {
		password = str
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if str := os.Getenv("MINIO_REGION_NAME"); str != "" {
		region = str
	}
	if str := u.Query().Get("region"); str != "" {
		region = str
	}

	sessionToken := os.Getenv("AWS_SESSION_TOKEN")
	if str := os.Getenv("MINIO_SESSION_TOKEN"); str != "" {
		sessionToken = str
	}
	if str := u.Query().Get("session_token"); str != "" {
		sessionToken = str
	}

	secureStr := os.Getenv("MINIO_SECURE")
	secureStr = u.Query().Get("secure")
	secure := false
	if secureStr != "" {
		secure, err = strconv.ParseBool(secureStr)
		if err != nil {
			return nil, err
		}
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(user, password, sessionToken),
		Secure: secure,
		Region: region,
	}

	return minio.New(endpoint, opts)
}
