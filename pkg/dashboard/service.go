package dashboard

// Repository provides access to relevant storage
type repository interface {
	ListAttributes() ([]string, error)
	RegisterWebhook(WebhookRequest, string) error
}

// Service provides dashboard operations
type Service interface {
	ListAttributes() ([]string, error)
	RegisterWebhook(WebhookRequest, string) error
}

type service struct {
	r repository
}

// NewService creates a dashboard service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// ListAttributes lists all observations in the Supertype ecosystem
func (s *service) ListAttributes() ([]string, error) {
	res, err := s.r.ListAttributes()
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
