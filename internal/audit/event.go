package audit

const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

type Event struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}
