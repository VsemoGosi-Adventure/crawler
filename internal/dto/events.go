package dto

import "time"

type CrawledEventMessage struct {
	ExternalSource   string      `json:"external_source"`
	ExternalSourceID string      `json:"external_source_id"`
	ExternalURL      string      `json:"external_url"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	StartsAt         time.Time   `json:"starts_at"`
	EndsAt           *time.Time  `json:"ends_at,omitempty"`
	MaxParticipants  *int        `json:"max_participants,omitempty"`
	JoinPolicy       string      `json:"join_policy"`
	Location         LocationDTO `json:"location"`
	Tags             []string    `json:"tags"`
}

type LocationDTO struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}
