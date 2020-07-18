package dynamo

// Observation defines an encrypted vendor observation
// TODO add IDs to observations
// TODO add vendor signing to observations
type Observation struct {
	Ciphertext string `json:"ciphertext"`
	Capsule    string `json:"capsule"`
	DateAdded  string `json:"dateAdded"`
}
