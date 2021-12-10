package clients

import (
	"math"
	"time"

	"github.com/mosuka/phalanx/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type IndexClient struct {
	proto.IndexClient
	conn *grpc.ClientConn
}

func NewIndexClient(address string) (*IndexClient, error) {
	return NewIndexClientWithTLS(address, "", "")
}

func NewIndexClientWithTLS(address string, certFile string, commonName string) (*IndexClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt64),
			grpc.MaxCallRecvMsgSize(math.MaxInt64),
		),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                1 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: true,
			},
		),
	}

	if certFile != "" && commonName != "" {
		creds, err := credentials.NewClientTLSFromFile(certFile, commonName)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(address, dialOpts...)
	if err != nil {
		return nil, err
	}
	client := proto.NewIndexClient(conn)

	return &IndexClient{
		IndexClient: client,
		conn:        conn,
	}, nil
}

func (c *IndexClient) Address() string {
	return c.conn.Target()
}

func (c *IndexClient) Close() error {
	return c.conn.Close()
}
