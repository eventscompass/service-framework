package pubsub

import (
	"time"
)

var (
	// Exchanges.
	EventsExchange string = "events"

	// Topics.
	EventCreatedTopic    string = "event.created"
	EventBookedTopic     string = "event.booked"
	LocationCreatedTopic string = "location.created"
)

// EventCreated is the payload for notifying for the creation of an event.
type EventCreated struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	LocationID string    `json:"location_id"`
	Start      time.Time `json:"start_time"`
	End        time.Time `json:"end_time"`
}

// LocationCreated is the payload for notifying for the creation of a location.
type LocationCreated struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EventBooked is the payload for notifying for the booking of an event.
type EventBooked struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}
