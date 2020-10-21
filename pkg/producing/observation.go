package producing

// ObservationRequest defines an encrypted vendor observation
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	Ciphertext  string `json:"ciphertext"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	SkHash      string `json:"skHash"`
}
