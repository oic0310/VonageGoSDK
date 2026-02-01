package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
)

const (
	// BaseURL is the Vonage Voice API base URL
	BaseURL = "https://api.nexmo.com"
)

// Client handles Vonage Voice API operations
type Client struct {
	baseURL      string
	phoneNumber  string
	jwtGenerator *vonage.JWTGenerator
	httpClient   *http.Client
}

// ClientOption is a functional option for configuring the voice client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL overrides the base URL (useful for testing)
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithPhoneNumber sets the caller phone number
func WithPhoneNumber(number string) ClientOption {
	return func(c *Client) {
		c.phoneNumber = number
	}
}

// NewClient creates a new Vonage Voice API client
func NewClient(jwtGenerator *vonage.JWTGenerator, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:      BaseURL,
		jwtGenerator: jwtGenerator,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewClientFromCredentials creates a new client from Vonage credentials
func NewClientFromCredentials(creds *vonage.Credentials, opts ...ClientOption) (*Client, error) {
	if !creds.HasApplication() {
		return nil, vonage.ErrNotConfigured
	}

	jwtGen := vonage.NewJWTGenerator(creds.AppID, creds.PrivateKey)
	allOpts := make([]ClientOption, 0, len(opts)+1)
	if creds.PhoneNumber != "" {
		allOpts = append(allOpts, WithPhoneNumber(creds.PhoneNumber))
	}
	allOpts = append(allOpts, opts...)

	return NewClient(jwtGen, allOpts...), nil
}

// PhoneNumber returns the configured phone number
func (c *Client) PhoneNumber() string {
	return c.phoneNumber
}

// ========================================
// Create Call
// ========================================

// CreateCall initiates a new outbound call
func (c *Client) CreateCall(ctx context.Context, opts CreateCallOptions) (*CreateCallResponse, error) {
	from := Endpoint{Type: EndpointTypePhone, Number: c.phoneNumber}
	if opts.From != nil {
		from = *opts.From
	}

	req := CreateCallRequest{
		To:   []Endpoint{opts.To},
		From: from,
	}

	// Use inline NCCO or answer URL
	if opts.InlineNCCO != nil {
		req.NCCO = opts.InlineNCCO
	} else {
		if opts.AnswerURL != "" {
			req.AnswerURL = []string{opts.AnswerURL}
		}
		if opts.AnswerMethod != "" {
			req.AnswerMethod = opts.AnswerMethod
		} else {
			req.AnswerMethod = "POST"
		}
	}

	if opts.EventURL != "" {
		req.EventURL = []string{opts.EventURL}
		if opts.EventMethod != "" {
			req.EventMethod = opts.EventMethod
		} else {
			req.EventMethod = "POST"
		}
	}

	return c.doCreateCall(ctx, req)
}

// CreateCallToPhone is a convenience method to call a phone number with answer/event URLs
func (c *Client) CreateCallToPhone(ctx context.Context, toNumber, answerURL, eventURL string) (*CreateCallResponse, error) {
	return c.CreateCall(ctx, CreateCallOptions{
		To:        PhoneEndpoint(toNumber),
		AnswerURL: answerURL,
		EventURL:  eventURL,
	})
}

// CreateCallWithNCCO is a convenience method to call with an inline NCCO
func (c *Client) CreateCallWithNCCO(ctx context.Context, toNumber string, ncco NCCO, eventURL string) (*CreateCallResponse, error) {
	return c.CreateCall(ctx, CreateCallOptions{
		To:         PhoneEndpoint(toNumber),
		InlineNCCO: ncco,
		EventURL:   eventURL,
	})
}

func (c *Client) doCreateCall(ctx context.Context, req CreateCallRequest) (*CreateCallResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/calls", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, vonage.NewError(resp.StatusCode, string(respBody))
	}

	var callResp CreateCallResponse
	if err := json.NewDecoder(resp.Body).Decode(&callResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Debug().
		Str("uuid", callResp.UUID).
		Str("status", callResp.Status).
		Msg("Call created")

	return &callResp, nil
}

// ========================================
// Get Call Info
// ========================================

// GetCallInfo retrieves information about a specific call
func (c *Client) GetCallInfo(ctx context.Context, callUUID string) (*CallInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/calls/%s", c.baseURL, callUUID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, vonage.NewError(resp.StatusCode, string(respBody))
	}

	var callInfo CallInfo
	if err := json.NewDecoder(resp.Body).Decode(&callInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &callInfo, nil
}

// ========================================
// Transfer Call
// ========================================

// TransferCall transfers an active call to a new NCCO URL
func (c *Client) TransferCall(ctx context.Context, callUUID, nccoURL string) error {
	req := TransferCallRequest{
		Action: "transfer",
		Destination: TransferDestination{
			Type: "ncco",
			URL:  []string{nccoURL},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	log.Debug().
		Str("callUUID", callUUID).
		Str("nccoURL", nccoURL).
		Msg("Call transferred")

	return nil
}

// ========================================
// Hangup Call
// ========================================

// HangupCall terminates an active call
func (c *Client) HangupCall(ctx context.Context, callUUID string) error {
	reqBody := map[string]string{"action": "hangup"}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	log.Debug().
		Str("callUUID", callUUID).
		Msg("Call hung up")

	return nil
}

// ========================================
// Mute / Unmute / Earmuff
// ========================================

// MuteCall mutes an active call
func (c *Client) MuteCall(ctx context.Context, callUUID string) error {
	return c.callAction(ctx, callUUID, "mute")
}

// UnmuteCall unmutes an active call
func (c *Client) UnmuteCall(ctx context.Context, callUUID string) error {
	return c.callAction(ctx, callUUID, "unmute")
}

// EarmuffCall earmuffs a call (recipient can't hear caller)
func (c *Client) EarmuffCall(ctx context.Context, callUUID string) error {
	return c.callAction(ctx, callUUID, "earmuff")
}

// UnearmuffCall removes earmuff from a call
func (c *Client) UnearmuffCall(ctx context.Context, callUUID string) error {
	return c.callAction(ctx, callUUID, "unearmuff")
}

func (c *Client) callAction(ctx context.Context, callUUID, action string) error {
	reqBody := map[string]string{"action": action}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	log.Debug().
		Str("callUUID", callUUID).
		Str("action", action).
		Msg("Call action executed")

	return nil
}

// ========================================
// Send DTMF
// ========================================

// SendDTMF sends DTMF tones to an active call
func (c *Client) SendDTMF(ctx context.Context, callUUID, digits string) error {
	reqBody := map[string]string{"digits": digits}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s/dtmf", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	return nil
}

// ========================================
// Play TTS into active call
// ========================================

// TalkIntoCall sends a TTS message into an active call
func (c *Client) TalkIntoCall(ctx context.Context, callUUID, text, voiceName string, loop int) error {
	reqBody := map[string]interface{}{
		"text":      text,
		"voice_name": voiceName,
		"loop":      loop,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s/talk", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	return nil
}

// StopTalk stops TTS in an active call
func (c *Client) StopTalk(ctx context.Context, callUUID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/v1/calls/%s/talk", c.baseURL, callUUID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	return nil
}

// ========================================
// Stream Audio into active call
// ========================================

// StreamIntoCall streams audio into an active call
func (c *Client) StreamIntoCall(ctx context.Context, callUUID string, streamURL string, loop int) error {
	reqBody := map[string]interface{}{
		"stream_url": []string{streamURL},
		"loop":       loop,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", fmt.Sprintf("%s/v1/calls/%s/stream", c.baseURL, callUUID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	return nil
}

// StopStream stops audio streaming in an active call
func (c *Client) StopStream(ctx context.Context, callUUID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/v1/calls/%s/stream", c.baseURL, callUUID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if err := c.setAuthHeaders(httpReq); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return vonage.NewError(resp.StatusCode, string(respBody))
	}

	return nil
}

// ========================================
// Auth helpers
// ========================================

func (c *Client) setAuthHeaders(req *http.Request) error {
	token, err := c.jwtGenerator.GenerateAPIJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}
