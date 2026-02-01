package messages

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
	// BaseURL is the Vonage Messages API base URL
	BaseURL = "https://api.nexmo.com"
)

// Client handles Vonage Messages API operations
type Client struct {
	baseURL      string
	phoneNumber  string
	jwtGenerator *vonage.JWTGenerator
	httpClient   *http.Client
}

// ClientOption is a functional option for configuring the messages client
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

// WithPhoneNumber sets the default sender phone number
func WithPhoneNumber(number string) ClientOption {
	return func(c *Client) {
		c.phoneNumber = number
	}
}

// NewClient creates a new Vonage Messages API client
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

// PhoneNumber returns the configured default phone number
func (c *Client) PhoneNumber() string {
	return c.phoneNumber
}

// ========================================
// Send Message (Generic)
// ========================================

// Send sends a message using the Vonage Messages API
func (c *Client) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	// Apply default sender if not set
	if req.From == "" {
		req.From = c.phoneNumber
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewReader(body))
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

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(respBody)).
			Msg("Vonage Messages API error")
		return nil, vonage.NewError(resp.StatusCode, string(respBody))
	}

	var sendResp SendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sendResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Debug().
		Str("messageUUID", sendResp.MessageUUID).
		Str("to", req.To).
		Str("channel", string(req.Channel)).
		Msg("Message sent")

	return &sendResp, nil
}

// ========================================
// SMS Convenience Methods
// ========================================

// SendSMS sends an SMS text message
func (c *Client) SendSMS(ctx context.Context, to, text string, opts ...SendOption) (*SendResponse, error) {
	req := &SendRequest{
		To:          to,
		MessageType: MessageTypeText,
		Text:        text,
		Channel:     ChannelSMS,
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Send(ctx, req)
}

// SendSMSFrom sends an SMS text message from a specific number
func (c *Client) SendSMSFrom(ctx context.Context, from, to, text string, opts ...SendOption) (*SendResponse, error) {
	req := &SendRequest{
		From:        from,
		To:          to,
		MessageType: MessageTypeText,
		Text:        text,
		Channel:     ChannelSMS,
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Send(ctx, req)
}

// ========================================
// MMS Convenience Methods
// ========================================

// SendMMS sends an MMS image message
func (c *Client) SendMMS(ctx context.Context, to, imageURL, caption string, opts ...SendOption) (*SendResponse, error) {
	req := &SendRequest{
		To:          to,
		MessageType: MessageTypeImage,
		Channel:     ChannelMMS,
		Image: &MediaContent{
			URL:     imageURL,
			Caption: caption,
		},
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Send(ctx, req)
}

// ========================================
// WhatsApp Convenience Methods
// ========================================

// SendWhatsApp sends a WhatsApp text message
func (c *Client) SendWhatsApp(ctx context.Context, to, text string, opts ...SendOption) (*SendResponse, error) {
	req := &SendRequest{
		To:          to,
		MessageType: MessageTypeText,
		Text:        text,
		Channel:     ChannelWhatsApp,
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Send(ctx, req)
}

// SendWhatsAppImage sends a WhatsApp image message
func (c *Client) SendWhatsAppImage(ctx context.Context, to, imageURL, caption string, opts ...SendOption) (*SendResponse, error) {
	req := &SendRequest{
		To:          to,
		MessageType: MessageTypeImage,
		Channel:     ChannelWhatsApp,
		Image: &MediaContent{
			URL:     imageURL,
			Caption: caption,
		},
	}

	for _, opt := range opts {
		opt(req)
	}

	return c.Send(ctx, req)
}

// ========================================
// Message Builder (Fluent API)
// ========================================

// MessageBuilder provides a fluent API for building messages
type MessageBuilder struct {
	client *Client
	req    SendRequest
}

// NewMessage creates a new message builder
func (c *Client) NewMessage() *MessageBuilder {
	return &MessageBuilder{
		client: c,
		req: SendRequest{
			From: c.phoneNumber,
		},
	}
}

// To sets the recipient
func (b *MessageBuilder) To(to string) *MessageBuilder {
	b.req.To = to
	return b
}

// From sets the sender (overrides default)
func (b *MessageBuilder) From(from string) *MessageBuilder {
	b.req.From = from
	return b
}

// SMS sets the channel to SMS
func (b *MessageBuilder) SMS() *MessageBuilder {
	b.req.Channel = ChannelSMS
	return b
}

// WhatsApp sets the channel to WhatsApp
func (b *MessageBuilder) WhatsApp() *MessageBuilder {
	b.req.Channel = ChannelWhatsApp
	return b
}

// Viber sets the channel to Viber
func (b *MessageBuilder) Viber() *MessageBuilder {
	b.req.Channel = ChannelViber
	return b
}

// Text sets the message text
func (b *MessageBuilder) Text(text string) *MessageBuilder {
	b.req.MessageType = MessageTypeText
	b.req.Text = text
	return b
}

// Image sets the message image
func (b *MessageBuilder) Image(url, caption string) *MessageBuilder {
	b.req.MessageType = MessageTypeImage
	b.req.Image = &MediaContent{URL: url, Caption: caption}
	return b
}

// Audio sets the message audio
func (b *MessageBuilder) Audio(url string) *MessageBuilder {
	b.req.MessageType = MessageTypeAudio
	b.req.Audio = &MediaContent{URL: url}
	return b
}

// Video sets the message video
func (b *MessageBuilder) Video(url, caption string) *MessageBuilder {
	b.req.MessageType = MessageTypeVideo
	b.req.Video = &MediaContent{URL: url, Caption: caption}
	return b
}

// File sets the message file
func (b *MessageBuilder) File(url, name string) *MessageBuilder {
	b.req.MessageType = MessageTypeFile
	b.req.File = &MediaContent{URL: url, Name: name}
	return b
}

// ClientRef sets a client reference for tracking
func (b *MessageBuilder) ClientRef(ref string) *MessageBuilder {
	b.req.ClientRef = ref
	return b
}

// Send sends the message
func (b *MessageBuilder) Send(ctx context.Context) (*SendResponse, error) {
	return b.client.Send(ctx, &b.req)
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
