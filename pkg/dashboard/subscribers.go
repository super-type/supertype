package dashboard

// TODO this definitely shouldn't be in dashboard service as we move forward

// MasterBedroom is the high-level attribute for master bedroom
type MasterBedroom struct {
	Attribute   string   `json:"attribute"`
	Lights      Lights   `json:"lights"`
	Curtains    Curtains `json:"curatins"`
	Subscribers []string `json:"subscribers"`
}

// Lights is the mid-level attribute for lights
type Lights struct {
	Color       Color    `json:"color"`
	Status      Status   `json:"status"`
	Subscribers []string `json:"subscribers"`
}

// Curtains is the mid-level attribute for curtains
type Curtains struct {
	Status      Status   `json:"status"`
	Subscribers []string `json:"subscribers"`
}

// Color is the low-level attribute for color
type Color struct {
	Subscribers []string `json:"subscribers"`
}

// Status is the low-level attribute for status
type Status struct {
	Subscribers []string `json:"subscribers"`
}
