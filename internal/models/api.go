package models

// AlertsRequest represents the request body for fetching alerts
type AlertsRequest struct {
	// Dates is an array of date strings in YYYY-MM-DD format
	Dates []string `json:"dates"`

	// Optional filters
	Subtypes []string `json:"subtypes,omitempty"` // Filter by alert subtypes
	Streets  []string `json:"streets,omitempty"`  // Filter by street names
}

// AlertsResponse represents the response containing alerts and metadata
type AlertsResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Alerts  []PoliceAlert `json:"alerts"`
	Stats   ResponseStats `json:"stats"`
}

// ResponseStats provides statistics about the response
type ResponseStats struct {
	TotalAlerts      int      `json:"total_alerts"`
	DatesQueried     []string `json:"dates_queried"`
	SubtypesFiltered []string `json:"subtypes_filtered,omitempty"`
	StreetsFiltered  []string `json:"streets_filtered,omitempty"`
}
