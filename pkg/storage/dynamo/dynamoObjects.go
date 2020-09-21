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

// User defines a Supertype user
type User struct {
	Username    string `json:"username"`
	SupertypeID string `json:"supertypeID"`
}

// Observation is a DynamoDB observation
type Observation struct {
	Ciphertext string `json:"ciphertext"`
	// Capsule     string `json:"capsule"`
	CapsuleE    string `json:"capsuleE"`
	CapsuleV    string `json:"capsuleV"`
	CapsuleS    string `json:"capsuleS"`
	DateAdded   string `json:"dateAdded"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	SkHash      string `json:"skHash"`
}
