package dynamo

// Vendor defines a Supertype vendor
type Vendor struct {
	// VendorID     string              `json:"vendorId"`
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	PublicKey   string            `json:"pk"`
	SupertypeID string            `json:"supertypeID"`
	Connections map[string]string `json:"connections"`
}
