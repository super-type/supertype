package dashboard

// Repository provides access to relevant storage
type repository interface {
	ListObservations() ([]string, error)
}

// Service provides dashboard operations
type Service interface {
	ListObservations() ([]string, error)
}

type service struct {
	r repository
}

// NewService creates a dashboard service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// ListObservations lists all observations in the Supertype ecosystem
func (s *service) ListObservations() ([]string, error) {
	res, err := s.r.ListObservations()
	if err != nil {
		return nil, err
	}
	return res, nil
}
