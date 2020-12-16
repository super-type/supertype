package authenticating

// Vendor defines a Supertype vendor
type Vendor struct {
	FirstName      string               `json:"firstName"`
	LastName       string               `json:"lastName"`
	Email          string               `json:"email"`
	BusinessName   string               `json:"businessName"`
	Username       string               `json:"username"`
	Password       string               `json:"password"`
	PublicKey      string               `json:"pk"`
	SupertypeID    string               `json:"supertypeID"`
	Connections    map[string][2]string `json:"connections"`
	AccountBalance float32              `json:"accountBalance"`
}

// CreateVendor is a password-less struct to use when creating a new user
type CreateVendor struct {
	FirstName      string               `json:"firstName"`
	LastName       string               `json:"lastName"`
	Email          string               `json:"email"`
	BusinessName   string               `json:"businessName"`
	Username       string               `json:"username"`
	PublicKey      string               `json:"pk"`
	APIKeyHash     string               `json:"apiKeyHash"`
	SupertypeID    string               `json:"supertypeID"`
	Connections    map[string][2]string `json:"connections"`
	AccountBalance float32              `json:"accountBalance"`
}

// AuthenticatedVendor is a password-less struct including the JWT returned to the user
type AuthenticatedVendor struct {
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	Email          string  `json:"email"`
	BusinessName   string  `json:"businessName"`
	Username       string  `json:"username"`
	PublicKey      string  `json:"pk"`
	SupertypeID    string  `json:"supertypeID"`
	JWT            string  `json:"jwt"`
	AccountBalance float32 `json:"accountBalance"`
}

// AuthenticatedVendorFirstLogin is what's returned to the user only on first login
type AuthenticatedVendorFirstLogin struct {
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	Email          string  `json:"email"`
	BusinessName   string  `json:"businessName"`
	Username       string  `json:"username"`
	PublicKey      string  `json:"pk"`
	PrivateKey     string  `json:"sk"`
	SupertypeID    string  `json:"supertypeID"`
	JWT            string  `json:"jwt"`
	AccountBalance float32 `json:"accountBalance"`
}
