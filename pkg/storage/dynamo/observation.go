package dynamo

// Observation is a DynamoDB observation
type Observation struct {
	Ciphertext  string `json:"ciphertext"`
	Capsule     string `json:"capsule"`
	DateAdded   string `json:"dateAdded"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
}
