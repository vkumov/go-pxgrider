package server

import (
	"context"
	"errors"

	pb "github.com/vkumov/go-pxgrider/pkg"
)

func (s *server) GetConnectionLogs(ctx context.Context, req *pb.GetConnectionLogsRequest) (*pb.GetConnectionLogsResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()

	logs, err := c.GetLogs(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := c.GetLogsCount(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetConnectionLogsResponse{
		ConnectionLogs: logs.ToProto(),
		Total:          total,
		Limit:          limit,
		Offset:         offset,
	}, nil
}

func (s *server) DeleteConnectionLogs(ctx context.Context, req *pb.DeleteConnectionLogsRequest) (*pb.DeleteConnectionLogsResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	var deleted int64

	what := req.GetWhat()
	switch what := what.(type) {
	case *pb.DeleteConnectionLogsRequest_All:
		deleted, err = c.DeleteAllLogs(ctx)
	case *pb.DeleteConnectionLogsRequest_LogIds:
		deleted, err = c.DeleteLogs(ctx, what.LogIds.GetIds())
	default:
		return nil, errors.New("unknown what to delete")
	}

	return &pb.DeleteConnectionLogsResponse{Deleted: deleted}, nil
}
