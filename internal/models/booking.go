package models

import "time"

type Booking struct {
	ID            int       `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Gender        string    `json:"gender"`
	Birthday      time.Time `json:"birthday"`
	LaunchpadID   string    `json:"launchpad_id"`
	DestinationID int64     `json:"destination_id"`
	LaunchDate    time.Time `json:"launch_date"`
}
