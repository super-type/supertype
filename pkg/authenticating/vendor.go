package authenticating

// Vendor defines a Supertype vendor
type Vendor struct {
	// VendorID     string              `json:"vendorId"`
	FirstName   string               `json:"firstName"`
	LastName    string               `json:"lastName"`
	Username    string               `json:"username"`
	Password    string               `json:"password"`
	PublicKey   string               `json:"pk"`
	SupertypeID string               `json:"supertypeID"`
	Connections map[string][2]string `json:"connections"`
}

// CreateVendor is a password-less struct to use when creating a new user
type CreateVendor struct {
	FirstName   string               `json:"firstName"`
	LastName    string               `json:"lastName"`
	Username    string               `json:"username"`
	PublicKey   string               `json:"pk"`
	SupertypeID string               `json:"supertypeID"`
	Connections map[string][2]string `json:"connections"`
}

// AuthenticatedVendor is a password-less struct including the JWT returned to the user
type AuthenticatedVendor struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	JWT         string `json:"jwt"`
}

// AuthenticatedVendorFirstLogin is what's returned to the user only on first login
type AuthenticatedVendorFirstLogin struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	PublicKey   string `json:"pk"`
	PrivateKey  string `json:"sk"`
	SupertypeID string `json:"supertypeID"`
	JWT         string `json:"jwt"`
}
