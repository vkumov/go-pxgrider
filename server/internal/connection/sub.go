package connection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	gopxgrid "github.com/vkumov/go-pxgrid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	pb "github.com/vkumov/go-pxgrider/pkg"
	"github.com/vkumov/go-pxgrider/server/internal/db/models"
)

type (
	TopicName string

	Subscription struct {
		PubSub      string
		Destination string
		Service     string
		Topic       string

		s   *gopxgrid.Subscription[any]
		ps  gopxgrid.PubSub
		log *zerolog.Logger
	}

	SubscriptionSlice []*Subscription
)

var (
	ErrSubNotInitialized           = errors.New("subscription is not initialized")
	ErrPubSubServiceNotInitialized = errors.New("pubsub service is not initialized")
)

func (c *Connection) Subscribe(ctx context.Context, sname string, topic TopicName) (*Subscription, error) {
	cnsm := c.pxCnsm.Load()
	if cnsm == nil {
		return nil, fmt.Errorf("pxgrid consumer is not initialized")
	}

	service, err := c.normalizeServiceName(sname)
	if err != nil {
		return nil, err
	}

	var svc gopxgrid.PxGridService

	c.log.Debug().Str("service", string(service)).Str("topic", string(topic)).Msg("Getting service by name for subscription")
	svc, err = c.getServiceByName(sname)
	if err != nil {
		return nil, err
	}

	c.log.Debug().Str("service", string(service)).Str("topic", string(topic)).Msg("Subscribing to topic")
	rs, err := svc.On(string(topic)).Subscribe(ctx)
	if err != nil {
		c.log.Error().Err(err).Msg("Failed to subscribe")
		return nil, err
	}
	c.log.Info().Str("service", string(service)).Str("topic", string(topic)).Msg("Subscribed")

	logger := c.log.With().Str("service", string(service)).Str("topic", string(topic)).Logger()

	s := &Subscription{
		s:           rs,
		Destination: rs.Destination(),
		PubSub:      rs.PubSubService,
		Service:     string(service),
		Topic:       string(topic),
		ps:          cnsm.PubSub(rs.PubSubService),
		log:         &logger,
	}

	c.startConsuming(topic, s)

	c.storeSubscription(service, topic, s)

	return s, nil
}

func (c *Connection) Unsubscribe(sname string, topic TopicName) error {
	service, err := c.normalizeServiceName(sname)
	if err != nil {
		return err
	}

	s := c.FindSubscription(sname, topic)
	if err := s.s.Unsubscribe(); err != nil {
		c.log.Error().Err(err).Msg("Failed to unsubscribe")
	}
	c.log.Info().Str("service", string(service)).Str("topic", string(topic)).Msg("Unsubscribed")

	c.lock.Lock()
	delete(c.topics[service], topic)
	c.lock.Unlock()

	return nil
}

func (c *Connection) FindSubscription(sname string, topic TopicName) *Subscription {
	service, err := c.normalizeServiceName(sname)
	if err != nil {
		return nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.topics[service]; !ok {
		return nil
	}

	if s, ok := c.topics[service][topic]; ok {
		return s
	}

	return nil
}

func (c *Connection) AllSubscriptions() SubscriptionSlice {
	c.lock.Lock()
	defer c.lock.Unlock()

	subs := make(SubscriptionSlice, 0)
	for _, v := range c.topics {
		for _, s := range v {
			subs = append(subs, s)
		}
	}

	return subs
}

func (c *Connection) storeSubscription(service ServiceName, topic TopicName, s *Subscription) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.topics == nil {
		c.topics = make(map[ServiceName]map[TopicName]*Subscription)
	}

	if _, ok := c.topics[service]; !ok {
		c.topics[service] = make(map[TopicName]*Subscription)
	}

	c.topics[service][topic] = s
}

func (c *Connection) startConsuming(topic TopicName, s *Subscription) {
	go func(dataChan <-chan *gopxgrid.Message[any], clid string) {
		for data := range dataChan {
			if data.Err != nil {
				s.log.Error().Err(data.Err).Msg("Failed to read message")
				continue
			}
			if data.UnmarshalError != nil {
				s.log.Error().Err(data.UnmarshalError).Msg("Failed to unmarshal message")
				continue
			}

			s.log.Debug().Str("topic", string(topic)).Msg("Received message")

			m := models.Message{
				Client:    clid,
				Topic:     string(topic),
				Viewed:    null.BoolFrom(false),
				Timestamp: null.TimeFrom(time.Now()),
			}
			if data.Body != nil {
				if err := m.Message.Marshal(data.Body); err != nil {
					log.Error().Err(err).Msg("Failed to marshal message")
					continue
				}
			} else {
				m.Message.SetValid(data.Message.Body)
			}

			db := c.db.Load()
			if err := m.Insert(context.Background(), db, boil.Infer()); err != nil {
				log.Error().Err(err).Msg("Failed to insert message")
			}
		}
	}(s.s.C, c.id)
}

func (s *Subscription) Nodes() ([]string, error) {
	if s == nil {
		return nil, ErrSubNotInitialized
	}
	if s.ps == nil {
		return nil, ErrPubSubServiceNotInitialized
	}

	oNodes := s.ps.Nodes()
	nodes := make([]string, 0, len(oNodes))
	for _, n := range oNodes {
		nodes = append(nodes, n.NodeName)
	}

	return nodes, nil
}

// Subscription is a JSON Marshaller
func (s *Subscription) MarshalJSON() ([]byte, error) {
	nodes, err := s.Nodes()
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"nodes":       nodes,
		"pubsub":      s.PubSub,
		"destination": s.Destination,
		"service":     s.Service,
		"topic":       s.Topic,
	})
}

// Subscription is a JSON Unmarshaller
func (s *Subscription) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if v, ok := m["pubsub"].(string); ok {
		s.PubSub = v
	}

	if v, ok := m["destination"].(string); ok {
		s.Destination = v
	}

	if v, ok := m["service"].(string); ok {
		s.Service = v
	}

	if v, ok := m["topic"].(string); ok {
		s.Topic = v
	}

	return nil
}

func (s *Subscription) ToProto() *pb.Subscription {
	if s == nil {
		return nil
	}

	nodes, err := s.Nodes()
	if err != nil && !errors.Is(err, ErrPubSubServiceNotInitialized) {
		s.log.Error().Err(err).Msg("Failed to get nodes")
	}

	connected := false
	if s.s != nil {
		connected = s.s.Active()
	}

	return &pb.Subscription{
		Pubsub:      s.PubSub,
		Destination: s.Destination,
		Connected:   connected,
		Nodes:       nodes,
		Service:     s.Service,
		Topic:       s.Topic,
	}
}

func (s SubscriptionSlice) ToProto() []*pb.Subscription {
	if s == nil {
		return nil
	}

	p := make([]*pb.Subscription, 0, len(s))
	for _, sub := range s {
		p = append(p, sub.ToProto())
	}

	return p
}
