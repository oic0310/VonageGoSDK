package messages

import "time"

// ========================================
// Channel Types
// ========================================

// Channel represents a messaging channel
type Channel string

const (
	ChannelSMS       Channel = "sms"
	ChannelMMS       Channel = "mms"
	ChannelWhatsApp  Channel = "whatsapp"
	ChannelViber     Channel = "viber_service"
	ChannelMessenger Channel = "messenger"
)

// MessageType represents the type of message content
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeFile     MessageType = "file"
	MessageTypeCustom   MessageType = "custom"
	MessageTypeTemplate MessageType = "template"
)

// ========================================
// Send Message
// ========================================

// SendRequest represents the Vonage Messages API request body
type SendRequest struct {
	From        string      `json:"from"`
	To          string      `json:"to"`
	MessageType MessageType `json:"message_type"`
	Text        string      `json:"text,omitempty"`
	Channel     Channel     `json:"channel"`

	// MMS / WhatsApp / Rich content
	Image    *MediaContent `json:"image,omitempty"`
	Audio    *MediaContent `json:"audio,omitempty"`
	Video    *MediaContent `json:"video,omitempty"`
	File     *MediaContent `json:"file,omitempty"`

	// WhatsApp specific
	WhatsApp *WhatsAppOptions `json:"whatsapp,omitempty"`

	// Client reference (for matching status webhooks)
	ClientRef string `json:"client_ref,omitempty"`

	// Webhook URL override (per-message)
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookVersion string `json:"webhook_version,omitempty"`
}

// SendResponse represents the Vonage Messages API response
type SendResponse struct {
	MessageUUID string `json:"message_uuid"`
}

// MediaContent represents media content for MMS/WhatsApp messages
type MediaContent struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Name    string `json:"name,omitempty"`
}

// WhatsAppOptions contains WhatsApp-specific message options
type WhatsAppOptions struct {
	Policy   string            `json:"policy,omitempty"`
	Locale   string            `json:"locale,omitempty"`
	Template *WhatsAppTemplate `json:"template,omitempty"`
}

// WhatsAppTemplate represents a WhatsApp template message
type WhatsAppTemplate struct {
	Name       string                    `json:"name"`
	Parameters []WhatsAppTemplateParam   `json:"parameters,omitempty"`
}

// WhatsAppTemplateParam represents a template parameter
type WhatsAppTemplateParam struct {
	Default string `json:"default"`
}

// ========================================
// Inbound Message (Webhook)
// ========================================

// InboundMessage represents a Vonage inbound message webhook payload
// This is the Messages API v1 format
type InboundMessage struct {
	MessageUUID string    `json:"message_uuid"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Timestamp   time.Time `json:"timestamp"`
	Channel     Channel   `json:"channel"`
	MessageType string    `json:"message_type"`

	// Text content
	Text string `json:"text,omitempty"`

	// Media content
	Image *InboundMedia `json:"image,omitempty"`
	Audio *InboundMedia `json:"audio,omitempty"`
	Video *InboundMedia `json:"video,omitempty"`
	File  *InboundMedia `json:"file,omitempty"`
}

// InboundMedia represents media in an inbound message
type InboundMedia struct {
	URL     string `json:"url"`
	Caption string `json:"caption,omitempty"`
	Name    string `json:"name,omitempty"`
}

// InboundSMS represents a legacy inbound SMS webhook payload
// This is the older SMS API format still used by some Vonage configurations
type InboundSMS struct {
	MSISDN    string `json:"msisdn"`
	To        string `json:"to"`
	MessageID string `json:"messageId"`
	Text      string `json:"text"`
	Timestamp string `json:"message-timestamp"`
	Type      string `json:"type,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
}

// ToInboundMessage converts a legacy InboundSMS to the unified InboundMessage format
func (s *InboundSMS) ToInboundMessage() *InboundMessage {
	return &InboundMessage{
		MessageUUID: s.MessageID,
		From:        s.MSISDN,
		To:          s.To,
		Channel:     ChannelSMS,
		MessageType: "text",
		Text:        s.Text,
	}
}

// ========================================
// Message Status (Webhook)
// ========================================

// MessageStatus represents a Vonage message status webhook payload
type MessageStatus struct {
	MessageUUID string    `json:"message_uuid"`
	To          string    `json:"to"`
	From        string    `json:"from"`
	Timestamp   time.Time `json:"timestamp"`
	Status      Status    `json:"status"`
	Channel     Channel   `json:"channel,omitempty"`
	Error       *Error    `json:"error,omitempty"`
	Usage       *Usage    `json:"usage,omitempty"`
	ClientRef   string    `json:"client_ref,omitempty"`
}

// Status represents a message delivery status
type Status string

const (
	StatusSubmitted Status = "submitted"
	StatusDelivered Status = "delivered"
	StatusRead      Status = "read"
	StatusRejected  Status = "rejected"
	StatusFailed    Status = "failed"
)

// IsDelivered returns true if the message was delivered
func (s Status) IsDelivered() bool {
	return s == StatusDelivered || s == StatusRead
}

// IsFailed returns true if the message failed
func (s Status) IsFailed() bool {
	return s == StatusRejected || s == StatusFailed
}

// IsTerminal returns true if the status is a terminal state
func (s Status) IsTerminal() bool {
	return s.IsDelivered() || s.IsFailed()
}

// Error represents an error in a status webhook
type Error struct {
	Type    string `json:"type,omitempty"`
	Title   string `json:"title,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Usage represents message usage/pricing information
type Usage struct {
	Currency string `json:"currency,omitempty"`
	Price    string `json:"price,omitempty"`
}

// ========================================
// Message Options (Functional Options for Send)
// ========================================

// SendOption is a functional option for configuring a message
type SendOption func(*SendRequest)

// WithClientRef sets a client reference for tracking
func WithClientRef(ref string) SendOption {
	return func(r *SendRequest) {
		r.ClientRef = ref
	}
}

// WithWebhookURL overrides the status webhook URL for this message
func WithWebhookURL(url string) SendOption {
	return func(r *SendRequest) {
		r.WebhookURL = url
	}
}
