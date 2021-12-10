package gateway

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

type IndexGateway struct {
	httpAddress string
	grpcAddress string

	certificateFile string
	keyFile         string

	corsAllowedMethods []string
	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logger *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	client *clients.IndexClient

	router *gin.Engine

	listener net.Listener
}

func NewIndexGatewayWithTLS(httpAddress string, grpcAddress string, certificateFile string, keyFile string, commonName string, corsAllowedMethods []string, corsAllowedOrigins []string, corsAllowedHeaders []string, logger *zap.Logger) (*IndexGateway, error) {
	gatewayLogger := logger.Named("gateway")

	client, err := clients.NewIndexClientWithTLS(grpcAddress, certificateFile, commonName)
	if err != nil {
		return nil, err
	}

	marshaler := marshaler.NewMarshaler()

	ctx, cancel := context.WithCancel(context.Background())

	router := gin.Default()
	router.Use(setClient(client))
	router.Use(setMarshaler(marshaler))
	router.Use(ginzap.Ginzap(gatewayLogger, time.RFC3339, true))
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
		logger.Error("failed to create index service", zap.Error(err))
		return nil, err
	}

	return &IndexGateway{
		httpAddress: httpAddress,
		grpcAddress: grpcAddress,

		certificateFile: certificateFile,
		keyFile:         keyFile,

		corsAllowedMethods: corsAllowedMethods,
		corsAllowedOrigins: corsAllowedOrigins,
		corsAllowedHeaders: corsAllowedHeaders,

		logger: gatewayLogger,

		ctx:    ctx,
		cancel: cancel,

		client: client,

		router: router,

		listener: listener,
	}, nil
}

func (g *IndexGateway) Start() error {
	go func() {
		if g.certificateFile == "" && g.keyFile == "" {
			_ = http.Serve(g.listener, g.router)
		} else {
			_ = http.ServeTLS(g.listener, g.router, g.certificateFile, g.keyFile)
		}
	}()

	return nil
}

func (g *IndexGateway) Stop() error {
	defer g.cancel()

	if err := g.listener.Close(); err != nil {
		g.logger.Error("failed to close listener", zap.Error(err), zap.String("http_address", g.listener.Addr().String()))
	}

	if err := g.client.Close(); err != nil {
		g.logger.Error("failed to close gRPC connection", zap.Error(err), zap.String("grpc_address", g.client.Address()))
	}

	return nil
}
