package consuming

// Repository provides access to relevant storage
type repository interface {
	Consume(ObservationRequest) (*[]ObservationResponse, error)
	Subscribe(WSObservationRequest) error
	GenerateConnectionID() (*string, error)
}

// Service provides consuming operations
type Service interface {
	Consume(ObservationRequest) (*[]ObservationResponse, error)
	Subscribe(WSObservationRequest) error
	GenerateConnectionID() (*string, error)
}

type service struct {
	r repository
}

// NewService creates a consuming service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Consume consumes encrypted data from Supertype and returns it to vendors
func (s *service) Consume(o ObservationRequest) (*[]ObservationResponse, error) {
	observation, err := s.r.Consume(o)
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

// GenerateConnectionID adds specified attributes to relevant Redis lists
func (s *service) GenerateConnectionID() (*string, error) {
	observation, err := s.r.GenerateConnectionID()
	if err != nil {
		return nil, err
	}
	return observation, err
}
