package clients

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mosuka/phalanx/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func NewEtcdClientWithUri(uri string) (*clientv3.Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "etcd" {
		return nil, errors.ErrInvalidUri
	}

	endpoints := strings.Split(os.Getenv("ETCD_ENDPOINTS"), ",")
	if str := u.Query().Get("endpoints"); str != "" {
		endpoints = strings.Split(str, ",")
	}

	return NewEtcdClient(endpoints)
}

func NewEtcdClient(endpoints []string) (*clientv3.Client, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	return clientv3.New(cfg)
}
