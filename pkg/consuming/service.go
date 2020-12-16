package consuming

// Repository provides access to relevant storage
type repository interface {
	Consume(ObservationRequest, string) (*[]ObservationResponse, error)
}

// Service provides consuming operations
type Service interface {
	Consume(ObservationRequest, string) (*[]ObservationResponse, error)
}

type service struct {
	r repository
}

// NewService creates a consuming service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Consume consumes encrypted data from Supertype and returns it to vendors
func (s *service) Consume(o ObservationRequest, apiKeyHash string) (*[]ObservationResponse, error) {
	observation, err := s.r.Consume(o, apiKeyHash)
	if err != nil {
		return nil, err
	}
	return observation, err
}
