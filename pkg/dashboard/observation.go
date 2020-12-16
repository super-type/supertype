package dashboard

// ObservationRequest defines an encrypted vendor observation request
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	SupertypeID string `json:"supertypeID"`
	PublicKey   string `json:"pk"`
}
