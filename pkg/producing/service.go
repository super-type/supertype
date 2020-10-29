package producing

import (
	httpUtil "github.com/super-type/supertype/pkg/http"
)

// Repository provides access to relevant storage
type repository interface {
	Produce(ObservationRequest) error
	Broadcast(BroadcastRequest, map[string]httpUtil.Pool) error
}

// Service provides producing operations
type Service interface {
	Produce(ObservationRequest) error
	Broadcast(BroadcastRequest, map[string]httpUtil.Pool) error
}

type service struct {
	r repository
}

// NewService creates a producing service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Produce produces encrypted data to Supertype
func (s *service) Produce(o ObservationRequest) error {
	err := s.r.Produce(o)
	if err != nil {
		return err
	}
	return nil
}

// Broadcast broadcasts newly-produced data to all subscribed WebSocket connections
func (s *service) Broadcast(b BroadcastRequest, poolMap map[string]httpUtil.Pool) error {
	err := s.r.Broadcast(b, poolMap)
	if err != nil {
		return err
	}
	return nil
}
