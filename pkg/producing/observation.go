package producing

// ObservationRequest defines an encrypted vendor observation
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	Ciphertext  string `json:"ciphertext"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	SkHash      string `json:"skHash"`
	IV          string `json:"iv"`
}

// BroadcastRequest defines an ObservationRequest with a given poolID
type BroadcastRequest struct {
	Attribute   string `json:"attribute"`
	Ciphertext  string `json:"ciphertext"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
	SkHash      string `json:"skHash"`
	IV          string `json:"iv"`
	PoolID      string `json:"poolID"`
}
