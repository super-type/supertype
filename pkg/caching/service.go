package caching

import (
	"context"

	"github.com/gorilla/websocket"
)

// Repository provides access to relevant storage
type repository interface {
	Subscribe(ctx context.Context, conn *websocket.Conn, channel string) (*[]byte, error)
	Publish(ctx context.Context, channel string, messages interface{})
}

// Service provides consuming operations
type Service interface {
	Subscribe(ctx context.Context, conn *websocket.Conn, channel string) (*[]byte, error)
	Publish(ctx context.Context, channel string, messages interface{})
}

type service struct {
	r repository
}

// NewService creates a consuming service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Subscribe adds specified attributes to relevant Redis lists
func (s *service) Subscribe(ctx context.Context, conn *websocket.Conn, channel string) (*[]byte, error) {
	message, err := s.r.Subscribe(ctx, conn, channel)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Publish publishes the given message to all consumers
func (s *service) Publish(ctx context.Context, channel string, messages interface{}) {
	s.r.Publish(ctx, channel, messages)
}
