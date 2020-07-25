package consuming

// ObservationRequest defines an encrypted vendor observation request
// TODO add IDs to observations
// TODO add vendor signing to observations
type ObservationRequest struct {
	Attribute   string `json:"attribute"`
	SupertypeID string `json:"supertypeID"`
	PublicKey   string `json:"pk"`
}

// ObservationResponse defines an encrypted vendor observation response
// TODO should we return the public key of the producer to the consumer? Or keep it anonymous?
type ObservationResponse struct {
	Ciphertext string `json:"ciphertext"`
	// Capsule              string    `json:"capsule"`
	CapsuleE             string    `json:"capsuleE"`
	CapsuleV             string    `json:"capsuleV"`
	CapsuleS             string    `json:"capsuleS"`
	DateAdded            string    `json:"dateAdded"`
	PublicKey            string    `json:"pk"`
	SupertypeID          string    `json:"supertypeID"`
	ReencryptionMetadata [2]string `json:"reencryptionMetadata"`
}
