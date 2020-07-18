package producing

// Observation defines an encrypted vendor observation
// TODO add IDs to observations
// TODO add vendor signing to observations
type Observation struct {
	Attribute  string `json:"attribute"`
	Ciphertext string `json:"ciphertext"`
	Capsule    string `json:"capsule"`
}
