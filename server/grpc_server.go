package server

import (
	"math"
	"net"
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mosuka/phalanx/metric"
	"github.com/mosuka/phalanx/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type GRPCIndexServer struct {
	grpcAddress  string
	grpcService  proto.IndexServer
	grpcServer   *grpc.Server
	listener     net.Listener
	certFile     string
	keyFile      string
	certHostname string
	logger       *zap.Logger
}

func NewGRPCIndexServer(grpcAddress string, certificateFile string, keyFile string, commonName string, indexService proto.IndexServer, logger *zap.Logger) (*GRPCIndexServer, error) {
	serverLogger := logger.Named("server")

	// Make the gRPC options.
	grpcOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt64),
		grpc.MaxSendMsgSize(math.MaxInt64),
		grpc.StreamInterceptor(
			grpcmiddleware.ChainStreamServer(
				metric.GrpcMetrics.StreamServerInterceptor(),
				grpczap.StreamServerInterceptor(serverLogger),
			),
		),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				metric.GrpcMetrics.UnaryServerInterceptor(),
				grpczap.UnaryServerInterceptor(serverLogger),
			),
		),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				//MaxConnectionIdle:     0,
				//MaxConnectionAge:      0,
				//MaxConnectionAgeGrace: 0,
				Time:    5 * time.Second,
				Timeout: 5 * time.Second,
			},
		),
	}

	// Make the certification.
	if certificateFile != "" && keyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(certificateFile, keyFile)
		if err != nil {
			serverLogger.Error(err.Error(), zap.String("certificate_file", certificateFile), zap.String("key_file", keyFile))
			return nil, err
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// Make the gRPC grpcServer.
	grpcServer := grpc.NewServer(
		grpcOpts...,
	)

	// Register the gRPC server and index service.
	proto.RegisterIndexServer(grpcServer, indexService)

	// Initialize all metrics.
	metric.GrpcMetrics.InitializeMetrics(grpcServer)
	grpc_prometheus.Register(grpcServer)

	// Make the listener.
	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		serverLogger.Error(err.Error(), zap.String("grpc_address", grpcAddress))
		return nil, err
	}

	return &GRPCIndexServer{
		grpcAddress:  grpcAddress,
		grpcService:  indexService,
		grpcServer:   grpcServer,
		listener:     listener,
		certFile:     certificateFile,
		keyFile:      keyFile,
		certHostname: commonName,
		logger:       serverLogger,
	}, nil
}

func (s *GRPCIndexServer) Start() error {
	go func() {
		err := s.grpcServer.Serve(s.listener)
		if err != nil {
			s.logger.Error(err.Error(), zap.String("address", s.grpcAddress))
		}
	}()

	return nil
}

func (s *GRPCIndexServer) Stop() error {
	s.grpcServer.GracefulStop()

	return nil
}
