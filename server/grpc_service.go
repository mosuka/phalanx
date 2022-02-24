package server

import (
	"bytes"
	"context"

	"github.com/mosuka/phalanx/metric"
	"github.com/mosuka/phalanx/proto"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCIndexService struct {
	proto.UnimplementedIndexServer

	indexService    *IndexService
	certificateFile string
	commonName      string
	logger          *zap.Logger
}

func NewGRPCIndexService(indexService *IndexService, certificateFile string, commonName string, logger *zap.Logger) (*GRPCIndexService, error) {
	serviceLogger := logger.Named("service")

	return &GRPCIndexService{
		indexService:    indexService,
		certificateFile: certificateFile,
		commonName:      commonName,
		logger:          serviceLogger,
	}, nil
}

func (s *GRPCIndexService) LivenessCheck(ctx context.Context, req *proto.LivenessCheckRequest) (*proto.LivenessCheckResponse, error) {
	resp := &proto.LivenessCheckResponse{}
	resp.State = proto.LivenessState_LIVENESS_STATE_ALIVE

	return resp, nil
}

func (s *GRPCIndexService) ReadinessCheck(ctx context.Context, req *proto.ReadinessCheckRequest) (*proto.ReadinessCheckResponse, error) {
	resp := &proto.ReadinessCheckResponse{}
	resp.State = proto.ReadinessState_READINESS_STATE_READY

	return resp, nil
}

func (s *GRPCIndexService) Metrics(ctx context.Context, req *proto.MetricsRequest) (*proto.MetricsResponse, error) {
	gather, err := metric.Registry.Gather()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	out := &bytes.Buffer{}
	for _, mf := range gather {
		if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
			s.logger.Warn(err.Error())
		}
	}

	resp := &proto.MetricsResponse{}
	resp.Metrics = out.Bytes()

	return resp, nil
}

func (s *GRPCIndexService) Cluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	resp, err := s.indexService.Cluster(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCIndexService) CreateIndex(ctx context.Context, req *proto.CreateIndexRequest) (*proto.CreateIndexResponse, error) {
	resp, err := s.indexService.CreateIndex(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCIndexService) DeleteIndex(ctx context.Context, req *proto.DeleteIndexRequest) (*proto.DeleteIndexResponse, error) {
	resp, err := s.indexService.DeleteIndex(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCIndexService) AddDocuments(ctx context.Context, req *proto.AddDocumentsRequest) (*proto.AddDocumentsResponse, error) {
	resp, err := s.indexService.AddDocuments(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCIndexService) DeleteDocuments(ctx context.Context, req *proto.DeleteDocumentsRequest) (*proto.DeleteDocumentsResponse, error) {
	resp, err := s.indexService.DeleteDocuments(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCIndexService) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	resp, err := s.indexService.Search(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
