package server

import (
	"context"

	pxgrider_proto "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/connection"
)

func (s *server) GetSubscription(ctx context.Context, req *pxgrider_proto.GetSubscriptionRequest) (*pxgrider_proto.GetSubscriptionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("GetSubscription")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	s.app.Log().Debug().Str("service", req.GetService()).Str("topic", req.GetTopic()).Msg("GetSubscription")
	sub := c.FindSubscription(req.Service, connection.TopicName(req.Topic))
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}

	return &pxgrider_proto.GetSubscriptionResponse{Subscription: sub.ToProto()}, nil
}

func (s *server) SubscribeConnection(ctx context.Context, req *pxgrider_proto.SubscribeConnectionRequest) (*pxgrider_proto.SubscribeConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("SubscribeConnection")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	s.app.Log().Debug().Str("service", req.GetService()).Str("topic", req.GetTopic()).Msg("Subscribe")
	sub, err := c.Subscribe(ctx, req.Service, connection.TopicName(req.Topic))
	if err != nil {
		return nil, err
	}

	return &pxgrider_proto.SubscribeConnectionResponse{Subscription: sub.ToProto()}, nil
}

func (s *server) UnsubscribeConnection(ctx context.Context, req *pxgrider_proto.UnsubscribeConnectionRequest) (*pxgrider_proto.UnsubscribeConnectionResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("UnsubscribeConnection")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	s.app.Log().Debug().Str("service", req.GetService()).Str("topic", req.GetTopic()).Msg("Unsubscribe")
	err = c.Unsubscribe(req.Service, connection.TopicName(req.Topic))
	if err != nil {
		return nil, err
	}

	return &pxgrider_proto.UnsubscribeConnectionResponse{}, nil
}

func (s *server) GetAllSubscriptions(ctx context.Context, req *pxgrider_proto.GetAllSubscriptionsRequest) (*pxgrider_proto.GetAllSubscriptionsResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("GetAllSubscriptions")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	sl := c.AllSubscriptions()
	resp := &pxgrider_proto.GetAllSubscriptionsResponse{
		Subscriptions: sl.ToProto(),
	}
	s.app.Log().Debug().Int("total", len(sl)).Msg("Subscriptions found")

	return resp, nil
}

func (s *server) GetConnectionTopics(ctx context.Context, req *pxgrider_proto.GetConnectionTopicsRequest) (*pxgrider_proto.GetConnectionTopicsResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("GetConnectionTopics")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	topics, err := c.GetAllTopics()
	if err != nil {
		return nil, err
	}

	resp := &pxgrider_proto.GetConnectionTopicsResponse{
		Topics: make(map[string]*pxgrider_proto.TopicsSlice, len(topics)),
	}

	for svc, t := range topics {
		resp.Topics[svc] = &pxgrider_proto.TopicsSlice{Topics: t}
	}
	s.app.Log().Debug().Int("total", len(resp.Topics)).Msg("Topics found")

	return resp, nil
}

func (s *server) GetServiceTopics(ctx context.Context, req *pxgrider_proto.GetServiceTopicsRequest) (*pxgrider_proto.GetServiceTopicsResponse, error) {
	s.app.Log().Debug().Str("uid", req.GetUser().Uid).Str("id", req.GetConnectionId()).Msg("GetServiceTopics")
	_, c, err := s.getUserConnection(ctx, req.GetUser().Uid, req.GetConnectionId())
	if err != nil {
		return nil, err
	}

	topics, err := c.GetTopicsOfService(req.GetServiceName())
	if err != nil {
		return nil, err
	}

	resp := &pxgrider_proto.GetServiceTopicsResponse{
		Topics: &pxgrider_proto.TopicsSlice{Topics: topics},
	}
	s.app.Log().Debug().Int("total", len(topics)).Msg("Topics found")

	return resp, nil
}
