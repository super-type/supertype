package producing

// Repository provides access to relevant storage
type repository interface {
	Produce(ObservationRequest) error
}

// Service provides producing operations
type Service interface {
	Produce(ObservationRequest) error
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
