package authenticating

// User defines a Supertype user
type User struct {
	Username    string `json:"username"`
	SupertypeID string `json:"supertypeID"`
	UserKey     string `json:"key"`
}

// UserWithVendors defines a user associated with Superytpe vendors
type UserWithVendors struct {
	Username    string   `json:"username"`
	SupertypeID string   `json:"supertypeID"`
	Vendors     []string `json:"vendors"`
}

// UserPassword is a password-less struct to use when handling user in any other
type UserPassword struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	SupertypeID string `json:"supertypeID"`
}
