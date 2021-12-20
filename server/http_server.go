package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/mosuka/phalanx/clients"
	"github.com/mosuka/phalanx/marshaler"
	"go.uber.org/zap"
)

const CorsMaxAge = 12 * time.Hour

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type HTTPIndexServer struct {
	httpAddress        string
	grpcAddress        string
	certificateFile    string
	keyFile            string
	corsAllowedMethods []string
	corsAllowedOrigins []string
	corsAllowedHeaders []string
	logger             *zap.Logger
	ctx                context.Context
	cancel             context.CancelFunc
	client             *clients.GRPCIndexClient
	router             *gin.Engine
	listener           net.Listener
}

func NewHTTPIndexServerWithTLS(httpAddress string, grpcAddress string, certificateFile string, keyFile string, commonName string, corsAllowedMethods []string, corsAllowedOrigins []string, corsAllowedHeaders []string, logger *zap.Logger) (*HTTPIndexServer, error) {
	httpLogger := logger.Named("http")

	client, err := clients.NewGRPCIndexClientWithTLS(grpcAddress, certificateFile, commonName)
	if err != nil {
		httpLogger.Error(err.Error(), zap.String("grpc_address", grpcAddress), zap.String("certificate_file", certificateFile), zap.String("common_name", commonName))
		return nil, err
	}

	marshaler := marshaler.NewMarshaler()

	ctx, cancel := context.WithCancel(context.Background())

	router := gin.Default()
	router.Use(setClient(client))
	router.Use(setMarshaler(marshaler))
	router.Use(ginzap.Ginzap(httpLogger, time.RFC3339, true))
	if len(corsAllowedOrigins) > 0 || len(corsAllowedMethods) > 0 || len(corsAllowedHeaders) > 0 {
		corsConfig := cors.Config{
			AllowOrigins:     corsAllowedOrigins,
			AllowMethods:     corsAllowedMethods,
			AllowHeaders:     corsAllowedHeaders,
			AllowCredentials: true,
			MaxAge:           CorsMaxAge,
		}
		router.Use(cors.New(corsConfig))
	}

	router.GET("/livez", livez)
	router.GET("/readyz", readyz)
	router.GET("/metrics", metrics)
	router.GET("/cluster", cluster)
	router.PUT("/v1/indexes/:index_name", putIndex)
	router.DELETE("/v1/indexes/:index_name", deleteIndex)
	router.PUT("/v1/indexes/:index_name/documents", putDocuments)
	router.DELETE("/v1/indexes/:index_name/documents", deleteDocuments)
	router.POST("/v1/indexes/:index_name/_search", search)

	listener, err := net.Listen("tcp", httpAddress)
	if err != nil {
		cancel()
		httpLogger.Error(err.Error(), zap.String("http_address", httpAddress))
		return nil, err
	}

	return &HTTPIndexServer{
		httpAddress:        httpAddress,
		grpcAddress:        grpcAddress,
		certificateFile:    certificateFile,
		keyFile:            keyFile,
		corsAllowedMethods: corsAllowedMethods,
		corsAllowedOrigins: corsAllowedOrigins,
		corsAllowedHeaders: corsAllowedHeaders,
		logger:             httpLogger,
		ctx:                ctx,
		cancel:             cancel,
		client:             client,
		router:             router,
		listener:           listener,
	}, nil
}

func (s *HTTPIndexServer) Start() error {
	go func() {
		if s.certificateFile == "" && s.keyFile == "" {
			_ = http.Serve(s.listener, s.router)
		} else {
			_ = http.ServeTLS(s.listener, s.router, s.certificateFile, s.keyFile)
		}
	}()

	return nil
}

func (s *HTTPIndexServer) Stop() error {
	defer s.cancel()

	if err := s.listener.Close(); err != nil {
		s.logger.Warn(err.Error())
	}

	if err := s.client.Close(); err != nil {
		s.logger.Error(err.Error())
	}

	return nil
}
