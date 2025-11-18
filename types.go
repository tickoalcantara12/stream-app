package stream

import "time"

type EventType string

const (
	EventStreamStart EventType = "stream.start"
	EventStreamEnd   EventType = "stream.end"
	EventCustom      EventType = "custom"
)

type Event struct {
	ID       string      `json:"id"`
	Type     EventType   `json:"type"`
	StreamID string      `json:"stream_id"`
	Data     interface{} `json:"data"`
	At       time.Time   `json:"at"`
}

type Metadata struct {
	Title string `json:"title"`
}
