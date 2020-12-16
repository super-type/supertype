package consuming

// ObservationRequest defines an encrypted vendor observation request
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	SupertypeID string `json:"supertypeID"`
	PublicKey   string `json:"pk"`
}

// ObservationResponse defines an encrypted vendor observation response
type ObservationResponse struct {
	Ciphertext  string `json:"ciphertext"`
	DateAdded   string `json:"dateAdded"`
	PublicKey   string `json:"pk"`
	SupertypeID string `json:"supertypeID"`
}
