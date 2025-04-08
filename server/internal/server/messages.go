package server

import (
	"context"
	"errors"

	pb "github.com/vkumov/go-pxgrider/pkg"
)

func (s *server) GetConnectionMessages(ctx context.Context, req *pb.GetConnectionMessagesRequest) (*pb.GetConnectionMessagesResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()

	msgs, err := c.GetMessages(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := c.GetMessagesCount(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetConnectionMessagesResponse{
		Messages: msgs.ToProto(),
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *server) MarkConnectionMessagesAsRead(ctx context.Context, req *pb.MarkConnectionMessagesAsReadRequest) (*pb.MarkConnectionMessagesAsReadResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	err = c.MarkMessages(ctx, req.GetMessageIds(), true)
	if err != nil {
		return nil, err
	}

	return &pb.MarkConnectionMessagesAsReadResponse{}, nil
}

func (s *server) DeleteConnectionMessages(ctx context.Context, req *pb.DeleteConnectionMessagesRequest) (*pb.DeleteConnectionMessagesResponse, error) {
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	var deleted int64

	what := req.GetWhat()
	switch what := what.(type) {
	case *pb.DeleteConnectionMessagesRequest_All:
		deleted, err = c.DeleteAllMessages(ctx)
	case *pb.DeleteConnectionMessagesRequest_MessageIds:
		deleted, err = c.DeleteMessages(ctx, what.MessageIds.GetIds())
	default:
		return nil, errors.New("unknown what to delete")
	}

	if err != nil {
		return nil, err
	}

	return &pb.DeleteConnectionMessagesResponse{Deleted: deleted}, nil
}
