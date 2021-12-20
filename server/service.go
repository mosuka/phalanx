package server

import (
	"bytes"
	"context"

	"github.com/mosuka/phalanx/index"
	"github.com/mosuka/phalanx/metric"
	"github.com/mosuka/phalanx/proto"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IndexService struct {
	indexManager    *index.Manager
	certificateFile string
	commonName      string
	logger          *zap.Logger
	proto.UnimplementedIndexServer
}

func NewIndexService(indexManager *index.Manager, certificateFile string, commonName string, logger *zap.Logger) (*IndexService, error) {
	serviceLogger := logger.Named("service")

	return &IndexService{
		indexManager:    indexManager,
		certificateFile: certificateFile,
		commonName:      commonName,
		logger:          serviceLogger,
	}, nil
}

func (s *IndexService) LivenessCheck(ctx context.Context, req *proto.LivenessCheckRequest) (*proto.LivenessCheckResponse, error) {
	resp := &proto.LivenessCheckResponse{}
	resp.State = proto.LivenessState_LIVENESS_STATE_ALIVE

	return resp, nil
}

func (s *IndexService) ReadinessCheck(ctx context.Context, req *proto.ReadinessCheckRequest) (*proto.ReadinessCheckResponse, error) {
	resp := &proto.ReadinessCheckResponse{}
	resp.State = proto.ReadinessState_READINESS_STATE_READY

	return resp, nil
}

func (s *IndexService) Metrics(ctx context.Context, req *proto.MetricsRequest) (*proto.MetricsResponse, error) {
	gather, err := metric.Registry.Gather()
	if err != nil {
		s.logger.Error("failed to get gather", zap.Error(err))
	}
	out := &bytes.Buffer{}
	for _, mf := range gather {
		if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
			s.logger.Warn("failed to parse metric family", zap.Error(err))
		}
	}

	resp := &proto.MetricsResponse{}
	resp.Metrics = out.Bytes()

	return resp, nil
}

func (s *IndexService) Cluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	resp, err := s.indexManager.Cluster(req)
	if err != nil {
		s.logger.Error("failed to get cluster information", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *IndexService) CreateIndex(ctx context.Context, req *proto.CreateIndexRequest) (*proto.CreateIndexResponse, error) {
	resp, err := s.indexManager.CreateIndex(req)
	if err != nil {
		s.logger.Error("failed to create index", zap.Error(err), zap.String("index_name", req.IndexName), zap.String("index_uri", req.IndexUri))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *IndexService) DeleteIndex(ctx context.Context, req *proto.DeleteIndexRequest) (*proto.DeleteIndexResponse, error) {
	resp, err := s.indexManager.DeleteIndex(req)
	if err != nil {
		s.logger.Error("failed to delete index", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *IndexService) AddDocuments(ctx context.Context, req *proto.AddDocumentsRequest) (*proto.AddDocumentsResponse, error) {
	resp, err := s.indexManager.AddDocuments(req)
	if err != nil {
		s.logger.Error("failed to add documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *IndexService) DeleteDocuments(ctx context.Context, req *proto.DeleteDocumentsRequest) (*proto.DeleteDocumentsResponse, error) {
	resp, err := s.indexManager.DeleteDocuments(req)
	if err != nil {
		s.logger.Error("failed to delete documents", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *IndexService) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	resp, err := s.indexManager.Search(req)
	if err != nil {
		s.logger.Error("failed to search index", zap.Error(err), zap.String("index_name", req.IndexName))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
