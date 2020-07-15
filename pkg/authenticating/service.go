package authenticating

import (
	"crypto/ecdsa"
)

// Repository provides access to relevant authentication storage
// ? Should we capitalize repository? It seems to be best practice to do so... but I don't see why?
type repository interface {
	CreateVendor(Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error)
	LoginVendor(Vendor) (*AuthenticatedVendor, error)
}

// Service provides authenticating operations
type Service interface {
	CreateVendor(Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error)
	LoginVendor(Vendor) (*AuthenticatedVendor, error)
}

type service struct {
	r repository
}

// NewService creates an auth service with the necessary dependencies
func NewService(r repository) Service {
	return &service{r}
}

// CreateVendor creates a vendor
func (s *service) CreateVendor(v Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error) {
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
