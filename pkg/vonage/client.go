package vonage

import (
	"net/http"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// API base URLs
	BaseURLREST  = "https://api.nexmo.com"
	BaseURLVideo = "https://video.api.vonage.com"
)

// Client is the main Vonage SDK client
type Client struct {
	credentials  *Credentials
	httpClient   *http.Client
	jwtGenerator *JWTGenerator

	// Sub-clients (lazy initialized)
	video *VideoClient
}

// ClientOption is a functional option for configuring the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Vonage client
func NewClient(credentials *Credentials, opts ...ClientOption) *Client {
	c := &Client{
		credentials: credentials,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	if credentials.HasApplication() {
		c.jwtGenerator = NewJWTGenerator(credentials.AppID, credentials.PrivateKey)
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Credentials returns the client's credentials
func (c *Client) Credentials() *Credentials {
	return c.credentials
}

// HTTPClient returns the HTTP client
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// JWTGenerator returns the JWT generator
func (c *Client) JWTGenerator() *JWTGenerator {
	return c.jwtGenerator
}

// VideoClient is a placeholder for the video sub-client
// The actual implementation is in the video package
type VideoClient struct{}

// Video returns the Video API client
// Note: This is a convenience method. For full Video API functionality,
// use the video package directly.
func (c *Client) Video() *VideoClient {
	if c.video == nil {
		c.video = &VideoClient{}
	}
	return c.video
}
