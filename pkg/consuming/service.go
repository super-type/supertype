package consuming

// Repository provides access to relevant storage
type repository interface {
	Consume(ObservationRequest) (*[]ObservationResponse, error)
	ConsumeWS(ObservationRequest) (*string, error)
}

// Service provides consuming operations
type Service interface {
	Consume(ObservationRequest) (*[]ObservationResponse, error)
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

// ConsumeWS establishes a WebSocket connection on the user-specified attributes
func (s *service) ConsumeWS(o ObservationRequest) (*string, error) { // TODO this will probably not just be a tring, but some JSON response once connection is established
	observation, err := s.r.ConsumeWS(o)
	if err != nil {
		return nil, err
	}
	return observation, err
}
