package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/vonatrigger/poc/internal/config"
	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/messages"
)

// VonageMessagesServiceV2 wraps the new SDK-based messages client
// This provides backward compatibility with the existing VonageService.SendSMS
type VonageMessagesServiceV2 struct {
	client      *messages.Client
	phoneNumber string
}

// NewVonageMessagesServiceV2 creates a new messages service using the SDK
func NewVonageMessagesServiceV2(cfg *config.Config, secrets VonageSecrets) (*VonageMessagesServiceV2, error) {
	creds, err := vonage.NewCredentials(
		vonage.WithApplication(secrets.AppID, secrets.PrivateKey),
		vonage.WithPhoneNumber(secrets.PhoneNumber),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	client, err := messages.NewClientFromCredentials(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create messages client: %w", err)
	}

	log.Info().
		Str("appID", secrets.AppID).
		Str("phoneNumber", secrets.PhoneNumber).
		Msg("Vonage Messages API V2 (SDK) configured")

	return &VonageMessagesServiceV2{
		client:      client,
		phoneNumber: secrets.PhoneNumber,
	}, nil
}

// SendSMS sends an SMS (backward compatible with existing VonageService.SendSMS)
func (s *VonageMessagesServiceV2) SendSMS(ctx context.Context, toNumber, text string) (*SendSMSResponse, error) {
	resp, err := s.client.SendSMS(ctx, toNumber, text)
	if err != nil {
		return nil, err
	}

	return &SendSMSResponse{
		MessageUUID: resp.MessageUUID,
	}, nil
}

// SendSMSWithRef sends an SMS with a client reference for tracking
func (s *VonageMessagesServiceV2) SendSMSWithRef(ctx context.Context, toNumber, text, clientRef string) (*SendSMSResponse, error) {
	resp, err := s.client.SendSMS(ctx, toNumber, text, messages.WithClientRef(clientRef))
	if err != nil {
		return nil, err
	}

	return &SendSMSResponse{
		MessageUUID: resp.MessageUUID,
	}, nil
}

// Client returns the underlying SDK messages client for advanced usage
func (s *VonageMessagesServiceV2) Client() *messages.Client {
	return s.client
}
