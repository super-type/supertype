package producing

// Repository provides access to relevant authentication storage
type repository interface {
	Produce(Observation) error
}

// Service provides producing operations
type Service interface {
	Produce(Observation) error
}

type service struct {
	r repository
}

// NewService creates a producing service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// Produce produces encrypted data to Supertype
func (s *service) Produce(o Observation) error {
	err := s.r.Produce(o)
	if err != nil {
		return err
	}
	return nil
}
