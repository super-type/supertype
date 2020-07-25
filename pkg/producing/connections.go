package producing

// MetadataResponse is the vendor metadata used to sync vendors on data production
type MetadataResponse struct {
	VendorConnections []string `json:"connections"`
	Vendors           []string `json:"vendors"`
}

// ReencryptionKeysRequest is sent when adding new re-encryption keys to a pre-existing vendor on produce
type ReencryptionKeysRequest struct {
	Connections map[string][]string `json:"connections"`
	PublicKey   string              `json:"pk"`
}
