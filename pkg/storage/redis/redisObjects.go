package redis

// Observation is a DynamoDB observation
type Observation struct {
	Ciphertext  string `json:"ciphertext"`
	DateAdded   string `json:"dateAdded"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	SkHash      string `json:"skHash"`
}
