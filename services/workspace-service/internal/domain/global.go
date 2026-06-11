package domain

type NotificationEventPayload struct {
	Channel string `json:"channel"`
	UserID  string `json:"userId,omitempty"`
	Title   string `json:"title,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
}
