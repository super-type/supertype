package producing

// MetadataResponse is the vendor metadata used to sync vendors on data production
type MetadataResponse struct {
	VendorConnections []string `json:"connections"`
	Vendors           []string `json:"vendors"`
}
