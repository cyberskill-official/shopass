package fcm

// Message represents the FCM HTTP v1 message structure.
type Message struct {
	Token        string            `json:"token"`
	Notification *MsgNotification  `json:"notification,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Webpush      *WebpushConfig    `json:"webpush,omitempty"`
	Android      *AndroidConfig    `json:"android,omitempty"`
}

type MsgNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type WebpushConfig struct {
	FCMOptions *FCMOptions `json:"fcm_options,omitempty"`
}

type FCMOptions struct {
	Link string `json:"link,omitempty"`
}

type AndroidConfig struct {
	Priority string `json:"priority,omitempty"` // "high" or "normal"
}
