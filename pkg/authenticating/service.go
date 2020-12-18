package authenticating

// Repository provides access to relevant authentication storage
// ? Should we capitalize repository? It seems to be best practice to do so... but I don't see why?
type repository interface {
	CreateVendor(Vendor) (*[2]string, error)
	LoginVendor(Vendor) (*AuthenticatedVendor, error)
	CreateUser(UserPassword) (*string, error)
	LoginUser(UserPassword) (*User, error)
	AuthorizedLoginUser(UserPassword, string) (*User, error)
}

// Service provides authenticating operations
type Service interface {
	CreateVendor(Vendor) (*[2]string, error)
	LoginVendor(Vendor) (*AuthenticatedVendor, error)
	CreateUser(UserPassword) (*string, error)
	LoginUser(UserPassword) (*User, error)
	AuthorizedLoginUser(UserPassword, string) (*User, error)
}

type service struct {
	r repository
}

// NewService creates an auth service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// CreateVendor creates a vendor
func (s *service) CreateVendor(v Vendor) (*[2]string, error) {
	result, err := s.r.CreateVendor(v)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LoginVendor logs in a vendor
func (s *service) LoginVendor(v Vendor) (*AuthenticatedVendor, error) {
	result, err := s.r.LoginVendor(v)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateUser creates a user
func (s *service) CreateUser(u UserPassword) (*string, error) {
	result, err := s.r.CreateUser(u)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// LoginUser creates a user
func (s *service) LoginUser(u UserPassword) (*User, error) {
	result, err := s.r.LoginUser(u)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AuthorizedLoginUser creates a user
func (s *service) AuthorizedLoginUser(u UserPassword, apiKey string) (*User, error) {
	result, err := s.r.AuthorizedLoginUser(u, apiKey)
	if err != nil {
		return nil, err
	}
	return result, nil
}
