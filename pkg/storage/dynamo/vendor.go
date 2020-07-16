package dynamo

// Vendor defines a Supertype vendor
type Vendor struct {
	// VendorID     string              `json:"vendorId"`
	FirstName      string               `json:"firstName"`
	LastName       string               `json:"lastName"`
	Email          string               `json:"email"`
	BusinessName   string               `json:"businessName"`
	Username       string               `json:"username"`
	PublicKey      string               `json:"pk"`
	SupertypeID    string               `json:"supertypeID"`
	Connections    map[string][2]string `json:"connections"`
	AccountBalance float32              `json:"accountBalance"`
}
