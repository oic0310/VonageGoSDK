package messages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ========================================
// Webhook Handlers
// ========================================

// InboundHandler is a function that handles inbound messages
type InboundHandler func(msg *InboundMessage) error

// StatusHandler is a function that handles message status updates
type StatusHandler func(status *MessageStatus) error

// WebhookHandler provides HTTP handler functions for Vonage webhooks
type WebhookHandler struct {
	onInbound InboundHandler
	onStatus  StatusHandler
	onLegacy  func(sms *InboundSMS) error
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

// OnInbound sets the handler for inbound messages (Messages API format)
func (h *WebhookHandler) OnInbound(handler InboundHandler) *WebhookHandler {
	h.onInbound = handler
	return h
}

// OnStatus sets the handler for message status updates
func (h *WebhookHandler) OnStatus(handler StatusHandler) *WebhookHandler {
	h.onStatus = handler
	return h
}

// OnLegacySMS sets the handler for legacy inbound SMS (older Vonage format)
func (h *WebhookHandler) OnLegacySMS(handler func(sms *InboundSMS) error) *WebhookHandler {
	h.onLegacy = handler
	return h
}

// HandleInbound returns an http.HandlerFunc for the inbound message webhook
func (h *WebhookHandler) HandleInbound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read inbound webhook body")
			w.WriteHeader(http.StatusOK) // Always 200 for webhooks
			return
		}
		defer r.Body.Close()

		// Try Messages API format first
		var msg InboundMessage
		if err := json.Unmarshal(body, &msg); err == nil && msg.MessageUUID != "" {
			if h.onInbound != nil {
				if err := h.onInbound(&msg); err != nil {
					log.Error().Err(err).Str("messageUUID", msg.MessageUUID).Msg("Error handling inbound message")
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Fall back to legacy SMS format
		var sms InboundSMS
		if err := json.Unmarshal(body, &sms); err == nil && sms.MSISDN != "" {
			if h.onLegacy != nil {
				if err := h.onLegacy(&sms); err != nil {
					log.Error().Err(err).Str("messageID", sms.MessageID).Msg("Error handling legacy inbound SMS")
				}
			} else if h.onInbound != nil {
				// Convert legacy to unified format
				unified := sms.ToInboundMessage()
				if err := h.onInbound(unified); err != nil {
					log.Error().Err(err).Str("from", sms.MSISDN).Msg("Error handling converted inbound SMS")
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Warn().Str("body", string(body)).Msg("Unknown inbound webhook format")
		w.WriteHeader(http.StatusOK)
	}
}

// HandleStatus returns an http.HandlerFunc for the message status webhook
func (h *WebhookHandler) HandleStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read status webhook body")
			w.WriteHeader(http.StatusOK)
			return
		}
		defer r.Body.Close()

		var status MessageStatus
		if err := json.Unmarshal(body, &status); err != nil {
			log.Warn().Str("body", string(body)).Msg("Failed to parse status webhook")
			w.WriteHeader(http.StatusOK)
			return
		}

		if h.onStatus != nil {
			if err := h.onStatus(&status); err != nil {
				log.Error().Err(err).
					Str("messageUUID", status.MessageUUID).
					Str("status", string(status.Status)).
					Msg("Error handling message status")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ========================================
// Parse Helpers (for use with Echo/Gin/etc)
// ========================================

// ParseInboundMessage parses an inbound message from a request body
func ParseInboundMessage(body []byte) (*InboundMessage, error) {
	// Try Messages API format
	var msg InboundMessage
	if err := json.Unmarshal(body, &msg); err == nil && msg.MessageUUID != "" {
		return &msg, nil
	}

	// Try legacy SMS format
	var sms InboundSMS
	if err := json.Unmarshal(body, &sms); err == nil && sms.MSISDN != "" {
		return sms.ToInboundMessage(), nil
	}

	return nil, fmt.Errorf("unknown inbound message format")
}

// ParseMessageStatus parses a message status from a request body
func ParseMessageStatus(body []byte) (*MessageStatus, error) {
	var status MessageStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse message status: %w", err)
	}
	return &status, nil
}
