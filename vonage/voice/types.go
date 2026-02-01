package voice

import "time"

// ========================================
// Call Types
// ========================================

// CallStatus represents the status of a call
type CallStatus string

const (
	CallStatusStarted   CallStatus = "started"
	CallStatusRinging   CallStatus = "ringing"
	CallStatusAnswered  CallStatus = "answered"
	CallStatusCompleted CallStatus = "completed"
	CallStatusFailed    CallStatus = "failed"
	CallStatusRejected  CallStatus = "rejected"
	CallStatusBusy      CallStatus = "busy"
	CallStatusCancelled CallStatus = "cancelled"
	CallStatusTimeout   CallStatus = "timeout"
)

// CallDirection represents the direction of a call
type CallDirection string

const (
	CallDirectionInbound  CallDirection = "inbound"
	CallDirectionOutbound CallDirection = "outbound"
)

// EndpointType represents the type of call endpoint
type EndpointType string

const (
	EndpointTypePhone     EndpointType = "phone"
	EndpointTypeSIP       EndpointType = "sip"
	EndpointTypeWebSocket EndpointType = "websocket"
	EndpointTypeVBC       EndpointType = "vbc"
)

// Endpoint represents a call endpoint (phone number, SIP, etc.)
type Endpoint struct {
	Type   EndpointType `json:"type"`
	Number string       `json:"number,omitempty"`
	URI    string       `json:"uri,omitempty"`
}

// PhoneEndpoint creates a phone endpoint
func PhoneEndpoint(number string) Endpoint {
	return Endpoint{
		Type:   EndpointTypePhone,
		Number: number,
	}
}

// SIPEndpoint creates a SIP endpoint
func SIPEndpoint(uri string) Endpoint {
	return Endpoint{
		Type: EndpointTypeSIP,
		URI:  uri,
	}
}

// ========================================
// Create Call
// ========================================

// CreateCallRequest represents the request to create a call
type CreateCallRequest struct {
	To           []Endpoint `json:"to"`
	From         Endpoint   `json:"from"`
	NCCO         NCCO       `json:"ncco,omitempty"`
	AnswerURL    []string   `json:"answer_url,omitempty"`
	AnswerMethod string     `json:"answer_method,omitempty"`
	EventURL     []string   `json:"event_url,omitempty"`
	EventMethod  string     `json:"event_method,omitempty"`
}

// CreateCallResponse represents the response from creating a call
type CreateCallResponse struct {
	UUID             string `json:"uuid"`
	Status           string `json:"status"`
	Direction        string `json:"direction"`
	ConversationUUID string `json:"conversation_uuid"`
}

// CreateCallOptions contains options for creating a call
type CreateCallOptions struct {
	// To is the destination endpoint
	To Endpoint
	// From is the caller endpoint (defaults to client's configured phone number)
	From *Endpoint
	// AnswerURL is the URL for Vonage to request NCCO from
	AnswerURL string
	// AnswerMethod is the HTTP method for the answer URL (default: POST)
	AnswerMethod string
	// EventURL is the URL for Vonage to send call events to
	EventURL string
	// EventMethod is the HTTP method for the event URL (default: POST)
	EventMethod string
	// InlineNCCO is an NCCO to use instead of an answer URL
	InlineNCCO NCCO
}

// ========================================
// Call Info
// ========================================

// CallInfo represents information about a call
type CallInfo struct {
	UUID             string        `json:"uuid"`
	Status           CallStatus    `json:"status"`
	Direction        CallDirection `json:"direction"`
	Rate             string        `json:"rate"`
	Price            string        `json:"price"`
	Duration         string        `json:"duration"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	ConversationUUID string        `json:"conversation_uuid"`
	Network          string        `json:"network,omitempty"`
	To               Endpoint      `json:"to,omitempty"`
	From             Endpoint      `json:"from,omitempty"`
}

// ========================================
// Transfer Call
// ========================================

// TransferCallRequest represents the request to transfer a call
type TransferCallRequest struct {
	Action      string              `json:"action"`
	Destination TransferDestination `json:"destination"`
}

// TransferDestination represents the transfer destination
type TransferDestination struct {
	Type string   `json:"type"`
	URL  []string `json:"url"`
}

// ========================================
// Call Event Webhook
// ========================================

// CallEvent represents a Vonage call event webhook payload
type CallEvent struct {
	UUID             string `json:"uuid"`
	ConversationUUID string `json:"conversation_uuid"`
	Status           string `json:"status"`
	Direction        string `json:"direction"`
	Timestamp        string `json:"timestamp"`
	From             string `json:"from,omitempty"`
	To               string `json:"to,omitempty"`
	Duration         string `json:"duration,omitempty"`
	Rate             string `json:"rate,omitempty"`
	Price            string `json:"price,omitempty"`
}

// IsTerminal returns true if the call event represents a terminal state
func (e *CallEvent) IsTerminal() bool {
	switch e.Status {
	case "completed", "failed", "rejected", "busy", "cancelled", "timeout":
		return true
	}
	return false
}

// ========================================
// ASR (Automatic Speech Recognition)
// ========================================

// ASRResult represents the Vonage ASR webhook payload
type ASRResult struct {
	Speech struct {
		TimeoutReason string      `json:"timeout_reason,omitempty"`
		Results       []ASRMatch  `json:"results,omitempty"`
	} `json:"speech,omitempty"`
	DTMF             string `json:"dtmf,omitempty"`
	UUID             string `json:"uuid"`
	ConversationUUID string `json:"conversation_uuid"`
	TimedOut         bool   `json:"timed_out"`
}

// ASRMatch represents a single ASR recognition result
type ASRMatch struct {
	Confidence string `json:"confidence"`
	Text       string `json:"text"`
}

// BestTranscript returns the best (first) speech recognition result
func (r *ASRResult) BestTranscript() string {
	if len(r.Speech.Results) > 0 {
		return r.Speech.Results[0].Text
	}
	return ""
}

// HasSpeech returns true if there are speech recognition results
func (r *ASRResult) HasSpeech() bool {
	return len(r.Speech.Results) > 0
}

// HasDTMF returns true if DTMF input was received
func (r *ASRResult) HasDTMF() bool {
	return r.DTMF != ""
}
