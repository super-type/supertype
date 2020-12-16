package dashboard

// Repository provides access to relevant storage
type repository interface {
	ListObservations(ObservationRequest, string) ([]string, error)
}

// Service provides dashboard operations
type Service interface {
	ListObservations(ObservationRequest, string) ([]string, error)
}

type service struct {
	r repository
}

// NewService creates a dashboard service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// ListObservations lists all observations in the Supertype ecosystem
func (s *service) ListObservations(o ObservationRequest, apiKeyHash string) ([]string, error) {
	res, err := s.r.ListObservations(o, apiKeyHash)
	if err != nil {
		return nil, err
	}
	return res, nil
}
