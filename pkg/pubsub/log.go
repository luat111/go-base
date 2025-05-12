package pubsub

type LogMsg struct {
	Mode          string `json:"mode"`
	CorrelationID string `json:"correlationID"`
	MessageValue  string `json:"messageValue"`
	Topic         string `json:"topic"`
	Host          string `json:"host"`
	Time          int64  `json:"time"`
}
