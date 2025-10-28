package events

type EmailEvent struct {
	To       string                 `json:"to"`
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
	Subject  string                 `json:"subject,omitempty"`
}
