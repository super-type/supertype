package dashboard

// Repository provides access to relevant storage
type repository interface {
	ListObservations(ObservationRequest, string) ([]string, error)
	RegisterWebhook(WebhookRequest, string) error
}

// Service provides dashboard operations
type Service interface {
	ListObservations(ObservationRequest, string) ([]string, error)
	RegisterWebhook(WebhookRequest, string) error
}

type service struct {
	r repository
}

// NewService creates a dashboard service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// ListObservations lists all observations in the Supertype ecosystem
func (s *service) ListObservations(o ObservationRequest, apiKey string) ([]string, error) {
	res, err := s.r.ListObservations(o, apiKey)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// RegisterWebhook creates a new webhook on a vendor's request
func (s *service) RegisterWebhook(webhookRequest WebhookRequest, apiKey string) error {
	err := s.r.RegisterWebhook(webhookRequest, apiKey)
	if err != nil {
		return err
	}
	return nil
}
