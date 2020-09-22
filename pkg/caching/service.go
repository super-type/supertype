package caching

// Repository provides access to relevant storage
type repository interface {
	GenerateConnectionID() (*string, error)
	Subscribe(WSObservationRequest) error
	GetSubscribers(ObservationRequest) (*[]string, error)
}

// Service provides consuming operations
type Service interface {
	GenerateConnectionID() (*string, error)
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

// GenerateConnectionID adds specified attributes to relevant Redis lists
func (s *service) GenerateConnectionID() (*string, error) {
	observation, err := s.r.GenerateConnectionID()
	if err != nil {
		return nil, err
	}
	return observation, err
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
