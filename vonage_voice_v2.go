package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/vonatrigger/poc/internal/config"
	"github.com/vonatrigger/poc/internal/model"
	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/voice"
)

// VonageServiceV2 wraps the new SDK-based voice client
// This provides backward compatibility with the existing VonageService interface
type VonageServiceV2 struct {
	voiceClient *voice.Client
	phoneNumber string
	webhookBase string
}

// NewVonageServiceV2 creates a new Vonage voice service using the SDK
func NewVonageServiceV2(cfg *config.Config, secrets VonageSecrets) (*VonageServiceV2, error) {
	creds, err := vonage.NewCredentials(
		vonage.WithApplication(secrets.AppID, secrets.PrivateKey),
		vonage.WithAPIKey(secrets.APIKey, secrets.APISecret),
		vonage.WithPhoneNumber(secrets.PhoneNumber),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %w", err)
	}

	voiceClient, err := voice.NewClientFromCredentials(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create voice client: %w", err)
	}

	log.Info().
		Str("appID", secrets.AppID).
		Str("phoneNumber", secrets.PhoneNumber).
		Msg("Vonage Voice API V2 (SDK) configured")

	return &VonageServiceV2{
		voiceClient: voiceClient,
		phoneNumber: secrets.PhoneNumber,
		webhookBase: cfg.WebhookBaseURL,
	}, nil
}

// ========================================
// Voice API (backward compatible)
// ========================================

// CreateCall initiates a call (backward compatible with existing VonageService)
func (s *VonageServiceV2) CreateCall(ctx context.Context, toNumber, answerPath, eventPath string) (*CreateCallResponse, error) {
	resp, err := s.voiceClient.CreateCall(ctx, voice.CreateCallOptions{
		To:           voice.PhoneEndpoint(toNumber),
		AnswerURL:    s.webhookBase + answerPath,
		AnswerMethod: "POST",
		EventURL:     s.webhookBase + eventPath,
		EventMethod:  "POST",
	})
	if err != nil {
		return nil, err
	}

	// Convert to existing response type
	return &CreateCallResponse{
		UUID:             resp.UUID,
		Status:           resp.Status,
		Direction:        resp.Direction,
		ConversationUUID: resp.ConversationUUID,
	}, nil
}

// GenerateNCCO generates NCCO for AI conversation (backward compatible)
func (s *VonageServiceV2) GenerateNCCO(text string, inputEventURL string) model.NCCO {
	ncco := voice.TalkAndInputJapanese(text, inputEventURL)
	return sdkNCCOToModel(ncco)
}

// GenerateNCCOWithStream generates NCCO that plays audio stream (backward compatible)
func (s *VonageServiceV2) GenerateNCCOWithStream(audioURL string, inputEventURL string) model.NCCO {
	ncco := voice.StreamAndInput(audioURL, inputEventURL, 1.5)
	return sdkNCCOToModel(ncco)
}

// GenerateNCCOHangup generates NCCO for ending a call (backward compatible)
func (s *VonageServiceV2) GenerateNCCOHangup(text string) model.NCCO {
	ncco := voice.TalkJapanese(text)
	return sdkNCCOToModel(ncco)
}

// TransferCall transfers an active call to a new NCCO
func (s *VonageServiceV2) TransferCall(ctx context.Context, callUUID, nccoURL string) error {
	return s.voiceClient.TransferCall(ctx, callUUID, nccoURL)
}

// GetCallInfo retrieves information about a specific call
func (s *VonageServiceV2) GetCallInfo(ctx context.Context, callUUID string) (*CallInfo, error) {
	info, err := s.voiceClient.GetCallInfo(ctx, callUUID)
	if err != nil {
		return nil, err
	}

	return &CallInfo{
		UUID:             info.UUID,
		Status:           string(info.Status),
		Direction:        string(info.Direction),
		Rate:             info.Rate,
		Price:            info.Price,
		Duration:         info.Duration,
		StartTime:        info.StartTime,
		EndTime:          info.EndTime,
		ConversationUUID: info.ConversationUUID,
	}, nil
}

// HangupCall terminates an active call
func (s *VonageServiceV2) HangupCall(ctx context.Context, callUUID string) error {
	return s.voiceClient.HangupCall(ctx, callUUID)
}

// ========================================
// Advanced SDK Access
// ========================================

// VoiceClient returns the underlying SDK voice client for advanced usage
func (s *VonageServiceV2) VoiceClient() *voice.Client {
	return s.voiceClient
}

// ========================================
// NCCO Conversion Helpers
// ========================================

// sdkNCCOToModel converts SDK NCCO to the existing model.NCCO format
func sdkNCCOToModel(sdkNCCO voice.NCCO) model.NCCO {
	ncco := make(model.NCCO, 0, len(sdkNCCO))

	for _, action := range sdkNCCO {
		modelAction := model.NCCOAction{
			Action:       action.ActionType,
			Text:         action.Text,
			VoiceName:    action.VoiceName,
			Language:     action.Language,
			Style:        action.Style,
			EventURL:     action.EventURL,
			EventMethod:  action.EventMethod,
			Type:         action.Type,
			EndOnSilence: action.EndOnSilence,
			StartTimeout: action.StartTimeout,
			MaxDuration:  action.MaxDuration,
			StreamURL:    action.StreamURL,
		}
		ncco = append(ncco, modelAction)
	}

	return ncco
}

// ModelNCCOToSDK converts model.NCCO to the SDK NCCO format
// Useful when migrating handlers to use SDK types directly
func ModelNCCOToSDK(modelNCCO model.NCCO) voice.NCCO {
	ncco := make(voice.NCCO, 0, len(modelNCCO))

	for _, action := range modelNCCO {
		sdkAction := voice.Action{
			ActionType:   action.Action,
			Text:         action.Text,
			VoiceName:    action.VoiceName,
			Language:     action.Language,
			Style:        action.Style,
			EventURL:     action.EventURL,
			EventMethod:  action.EventMethod,
			Type:         action.Type,
			EndOnSilence: action.EndOnSilence,
			StartTimeout: action.StartTimeout,
			MaxDuration:  action.MaxDuration,
			StreamURL:    action.StreamURL,
		}
		ncco = append(ncco, sdkAction)
	}

	return ncco
}
