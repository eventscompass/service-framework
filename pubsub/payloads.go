package pubsub

import (
	"time"

	"github.com/eventscompass/service-framework/service"
)

var (
	_ service.Payload = (*EventCreated)(nil)
	_ service.Payload = (*LocationCreated)(nil)
	_ service.Payload = (*EventBooked)(nil)
)

// EventCreated is the payload for notifying for the creation of an event.
type EventCreated struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	LocationID string    `json:"location_id"`
	Start      time.Time `json:"start_time"`
	End        time.Time `json:"end_time"`
}

// Topic implements the [service.Payload] interface.
func (*EventCreated) Topic() string {
	return eventCreatedTopic
}

// LocationCreated is the payload for notifying for the creation of a location.
type LocationCreated struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Topic implements the [service.Payload] interface.
func (*LocationCreated) Topic() string {
	return locationCreatedTopic
}

// EventBooked is the payload for notifying for the booking of an event.
type EventBooked struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}

// Topic implements the [service.Payload] interface.
func (*EventBooked) Topic() string {
	return eventBookedTopic
}

var (
	// topics
	eventCreatedTopic    string = "event.created"
	eventBookedTopic     string = "event.booked"
	locationCreatedTopic string = "location.created"
)
