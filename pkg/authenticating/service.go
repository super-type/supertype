package authenticating

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
)

// ErrVendorNotFound is used when a vendor is not found in the database
var ErrVendorNotFound = errors.New("Vendor not found")

// ErrVendorAlreadyExists is used when attempting to create an already-used username
var ErrVendorAlreadyExists = errors.New("Vendor already exists")

// Repository provides access to relevant authentication storage
// ? Should we capitalize repository? It seems to be best practice to do so... but I don't see why?
type repository interface {
	CreateVendor(Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error)
	LoginVendor(Vendor) error
}

// Service provides authenticating operations
type Service interface {
	CreateVendor(Vendor) (map[*ecdsa.PublicKey]*ecdsa.PrivateKey, error)
	LoginVendor(Vendor)
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
func (s *service) LoginVendor(v Vendor) {
	err := s.r.LoginVendor(v)
	if err != nil {
		fmt.Println(err)
	}
}
