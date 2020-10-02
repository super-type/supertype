package caching

// Repository provides access to relevant storage
type repository interface {
	Subscribe(WSObservationRequest) error
	GetSubscribers(ObservationRequest) (*[]string, error)
}

// Service provides consuming operations
type Service interface {
	Subscribe(WSObservationRequest) error
	GetSubscribers(ObservationRequest) (*[]string, error)
}

type service struct {
	r repository
}

// NewService creates a consuming service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Subscribe adds specified attributes to relevant Redis lists
func (s *service) Subscribe(o WSObservationRequest) error {
	err := s.r.Subscribe(o)
	if err != nil {
		return err
	}
	return nil
}

// GetSubscribers
func (s *service) GetSubscribers(o ObservationRequest) (*[]string, error) {
	subscribers, err := s.r.GetSubscribers(o)
	if err != nil {
		return nil, err
	}
	return subscribers, nil
}
