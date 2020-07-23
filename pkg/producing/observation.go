package producing

// ObservationRequest defines an encrypted vendor observation
// TODO add IDs to observations
// TODO add vendor signing to observations
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	Ciphertext  string `json:"ciphertext"`
	Capsule     string `json:"capsule"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
}
