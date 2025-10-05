package models

import (
	"time"

	"google.golang.org/genproto/googleapis/type/latlng"
)

// Location represents geographic coordinates
type Location struct {
	Latitude  float64 `json:"y" firestore:"latitude"`
	Longitude float64 `json:"x" firestore:"longitude"`
}

// Comment represents a user comment on an alert
type Comment struct {
	ReportMillis int64  `json:"reportMillis"`
	Text         string `json:"text"`
	IsThumbsUp   bool   `json:"isThumbsUp"`
}

// WazeAlert represents a single alert from Waze API
type WazeAlert struct {
	// Core identifiers
	UUID      string `json:"uuid" firestore:"uuid"`
	ID        string `json:"id,omitempty" firestore:"id,omitempty"`
	PubMillis int64  `json:"pubMillis" firestore:"pub_millis"`

	// Alert classification
	Type    string `json:"type" firestore:"type"`
	Subtype string `json:"subtype" firestore:"subtype"`

	// Location details
	Location Location `json:"location" firestore:"location"`
	Street   string   `json:"street,omitempty" firestore:"street,omitempty"`
	City     string   `json:"city,omitempty" firestore:"city,omitempty"`
	Country  string   `json:"country,omitempty" firestore:"country,omitempty"`

	// Road information
	RoadType   int `json:"roadType,omitempty" firestore:"road_type,omitempty"`
	FromNodeId int `json:"fromNodeId,omitempty" firestore:"from_node_id,omitempty"`
	ToNodeId   int `json:"toNodeId,omitempty" firestore:"to_node_id,omitempty"`

	// Reliability and confidence
	Reliability  int `json:"reliability,omitempty" firestore:"reliability,omitempty"`
	Confidence   int `json:"confidence,omitempty" firestore:"confidence,omitempty"`
	ReportRating int `json:"reportRating,omitempty" firestore:"report_rating,omitempty"`

	// Report details
	ReportBy                 string `json:"reportBy,omitempty" firestore:"report_by,omitempty"`
	ReportDescription        string `json:"reportDescription,omitempty" firestore:"report_description,omitempty"`
	ReportByMunicipalityUser string `json:"reportByMunicipalityUser,omitempty" firestore:"report_by_municipality_user,omitempty"`
	ReportMood               int    `json:"reportMood,omitempty" firestore:"report_mood,omitempty"`

	// Provider information
	Provider   string `json:"provider,omitempty" firestore:"provider,omitempty"`
	ProviderId string `json:"providerId,omitempty" firestore:"provider_id,omitempty"`

	// Community engagement
	NThumbsUp int       `json:"nThumbsUp,omitempty" firestore:"n_thumbs_up,omitempty"`
	NComments int       `json:"nComments,omitempty" firestore:"n_comments,omitempty"`
	Comments  []Comment `json:"comments,omitempty" firestore:"-"` // Not stored directly in WazeAlert

	// Additional fields
	Magvar         int    `json:"magvar,omitempty" firestore:"magvar,omitempty"`
	Speed          int    `json:"speed,omitempty" firestore:"speed,omitempty"`
	AdditionalInfo string `json:"additionalInfo,omitempty" firestore:"additional_info,omitempty"`
	WazeData       string `json:"wazeData,omitempty" firestore:"waze_data,omitempty"`
	Inscale        bool   `json:"inscale,omitempty" firestore:"inscale,omitempty"`
}

// PoliceAlert represents a tracked police alert with full lifecycle tracking
// This is what gets stored in Firestore for POLICE type alerts
type PoliceAlert struct {
	// Core alert data (from Waze API)
	UUID    string `firestore:"uuid"`
	ID      string `firestore:"id,omitempty"`
	Type    string `firestore:"type"`
	Subtype string `firestore:"subtype"`
	Street  string `firestore:"street,omitempty"`
	City    string `firestore:"city,omitempty"`
	Country string `firestore:"country,omitempty"`

	// Location as GeoPoint for geospatial queries
	LocationGeo *latlng.LatLng `firestore:"location_geo"`

	// Reliability metrics
	Reliability  int `firestore:"reliability,omitempty"`
	Confidence   int `firestore:"confidence,omitempty"`
	ReportRating int `firestore:"report_rating,omitempty"`

	// Time tracking (all as Firestore Timestamps)
	PublishTime time.Time `firestore:"publish_time"` // Converted from pubMillis
	ScrapeTime  time.Time `firestore:"scrape_time"`  // First time seen
	ExpireTime  time.Time `firestore:"expire_time"`  // Last time seen (assumes expired after)

	// Verification tracking
	LastVerificationTime *time.Time `firestore:"last_verification_time,omitempty"` // Latest comment timestamp

	// Duration tracking (in milliseconds for consistency with Waze)
	ActiveMillis           int64  `firestore:"active_millis"`                      // expireMillis - pubMillis
	LastVerificationMillis *int64 `firestore:"last_verification_millis,omitempty"` // Latest comment reportMillis

	// Community engagement tracking
	NThumbsUpInitial int `firestore:"n_thumbs_up_initial"` // Initial thumbs up count
	NThumbsUpLast    int `firestore:"n_thumbs_up_last"`    // Most recent thumbs up count

	// Raw data preservation
	RawDataInitial string `firestore:"raw_data_initial"` // First scrape JSON
	RawDataLast    string `firestore:"raw_data_last"`    // Most recent scrape JSON
}

// WazeGeoRSSResponse is the response from Waze API
type WazeGeoRSSResponse struct {
	Alerts []WazeAlert `json:"alerts"`
}

// WazeAPIResponse is an alias for compatibility
type WazeAPIResponse = WazeGeoRSSResponse

// ScrapingStats tracks scraping statistics
type ScrapingStats struct {
	TotalRequests     int       `json:"total_requests"`
	SuccessfulCalls   int       `json:"successful_calls"`
	FailedCalls       int       `json:"failed_calls"`
	TotalAlerts       int       `json:"total_alerts"`
	UniqueAlerts      int       `json:"unique_alerts"`
	LastSuccessfulRun time.Time `json:"last_successful_run"`
}
