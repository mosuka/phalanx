package clients

import (
	"math"
	"time"

	"github.com/mosuka/phalanx/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCIndexClient struct {
	proto.IndexClient
	conn *grpc.ClientConn
}

func NewGRPCIndexClient(address string) (*GRPCIndexClient, error) {
	return NewGRPCIndexClientWithTLS(address, "", "")
}

func NewGRPCIndexClientWithTLS(address string, certFile string, commonName string) (*GRPCIndexClient, error) {
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
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(address, dialOpts...)
	if err != nil {
		return nil, err
	}
	client := proto.NewIndexClient(conn)

	return &GRPCIndexClient{
		IndexClient: client,
		conn:        conn,
	}, nil
}

func (c *GRPCIndexClient) Address() string {
	return c.conn.Target()
}

func (c *GRPCIndexClient) Close() error {
	return c.conn.Close()
}
