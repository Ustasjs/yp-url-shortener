// Package audit provides an observer-based mechanism for recording audit events
// (such as URL shortening and redirects) and delivering them to configurable
// sinks like a file or an HTTP endpoint.
package audit

// Audit action identifiers used in Event.Action.
const (
	ActionShorten = "shorten" // creation of a short URL
	ActionFollow  = "follow"  // redirect to an original URL
)

// Event is a single audit record describing an action performed on a URL.
type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}
