package producing

// Repository provides access to relevant storage
type repository interface {
	Produce(ObservationRequest, string) error
}

// Service provides producing operations
type Service interface {
	Produce(ObservationRequest, string) error
}

type service struct {
	r repository
}

// NewService creates a producing service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Produce produces encrypted data to Supertype
func (s *service) Produce(o ObservationRequest, apiKeyHash string) error {
	err := s.r.Produce(o, apiKeyHash)
	if err != nil {
		return err
	}
	return nil
}
