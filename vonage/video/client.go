package video

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
)

const (
	// BaseURL is the Vonage Video API base URL
	BaseURL = "https://video.api.vonage.com"

	// DefaultSessionTTL is the default session time-to-live
	DefaultSessionTTL = 24 * time.Hour
)

// Client handles Vonage Video API operations
type Client struct {
	appID        string
	jwtGenerator *vonage.JWTGenerator
	httpClient   *http.Client

	// Session cache
	sessions map[string]*Session
	mu       sync.RWMutex
}

// ClientOption is a functional option for configuring the video client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Vonage Video API client
func NewClient(appID string, jwtGenerator *vonage.JWTGenerator, opts ...ClientOption) *Client {
	c := &Client{
		appID:        appID,
		jwtGenerator: jwtGenerator,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		sessions:     make(map[string]*Session),
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
	return NewClient(creds.AppID, jwtGen, opts...), nil
}

// IsConfigured returns true if the client has valid credentials
func (c *Client) IsConfigured() bool {
	return c.jwtGenerator != nil && c.appID != ""
}

// AppID returns the application ID
func (c *Client) AppID() string {
	return c.appID
}

// CreateSession creates a new video session
func (c *Client) CreateSession(opts *CreateSessionOptions) (*Session, error) {
	if !c.IsConfigured() {
		log.Warn().Msg("Vonage Video API not configured, using mock session")
		return c.createMockSession("")
	}

	session, err := c.createSessionViaAPI(opts)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create session via API, using mock session")
		return c.createMockSession("")
	}

	// Cache the session
	c.mu.Lock()
	c.sessions[session.SessionID] = session
	c.mu.Unlock()

	log.Info().Str("sessionID", session.SessionID).Msg("Created Vonage Video session")
	return session, nil
}

// CreateSessionForSpot creates a session associated with a specific spot
func (c *Client) CreateSessionForSpot(spotID string, opts *CreateSessionOptions) (*Session, error) {
	// Check cache first
	c.mu.RLock()
	for _, session := range c.sessions {
		if session.SpotID == spotID && session.IsValid() {
			c.mu.RUnlock()
			return session, nil
		}
	}
	c.mu.RUnlock()

	if !c.IsConfigured() {
		log.Warn().Msg("Vonage Video API not configured, using mock session")
		return c.createMockSession(spotID)
	}

	session, err := c.createSessionViaAPI(opts)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create session via API, using mock session")
		return c.createMockSession(spotID)
	}

	session.SpotID = spotID

	// Cache the session
	c.mu.Lock()
	c.sessions[session.SessionID] = session
	c.mu.Unlock()

	log.Info().
		Str("sessionID", session.SessionID).
		Str("spotID", spotID).
		Msg("Created Vonage Video session for spot")

	return session, nil
}

// createSessionViaAPI calls the Vonage Video API to create a session
func (c *Client) createSessionViaAPI(opts *CreateSessionOptions) (*Session, error) {
	apiJWT, err := c.jwtGenerator.GenerateAPIJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API JWT: %w", err)
	}

	apiURL := fmt.Sprintf("%s/session/create", BaseURL)

	// Build form data for session options
	formData := url.Values{}
	if opts != nil {
		if opts.Location != "" {
			formData.Set("location", opts.Location)
		}
		if opts.MediaMode != "" {
			formData.Set("p2p.preference", string(opts.MediaMode))
		}
		if opts.ArchiveMode != "" {
			formData.Set("archiveMode", string(opts.ArchiveMode))
		}
	}

	var req *http.Request
	if len(formData) > 0 {
		req, err = http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequest("POST", apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiJWT)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Str("url", apiURL).
			Msg("Vonage Video API error")
		return nil, vonage.NewError(resp.StatusCode, string(body))
	}

	// Response is an array of session objects
	var results []CreateSessionResponse
	if err := json.Unmarshal(body, &results); err != nil {
		// Try single object response
		var single CreateSessionResponse
		if err := json.Unmarshal(body, &single); err != nil {
			log.Error().
				Str("body", string(body)).
				Err(err).
				Msg("Failed to parse Vonage Video API response")
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		results = []CreateSessionResponse{single}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	return &Session{
		SessionID: results[0].SessionID,
		ProjectID: results[0].ProjectID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(DefaultSessionTTL),
	}, nil
}

// createMockSession creates a mock session for development/testing
func (c *Client) createMockSession(spotID string) (*Session, error) {
	appIDPrefix := "mock"
	if len(c.appID) >= 8 {
		appIDPrefix = c.appID[:8]
	}
	sessionID := fmt.Sprintf("mock_%s_%d", appIDPrefix, time.Now().UnixNano())

	log.Info().
		Str("sessionID", sessionID).
		Str("spotID", spotID).
		Msg("Created mock video session")

	session := &Session{
		SessionID: sessionID,
		SpotID:    spotID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(DefaultSessionTTL),
		IsMock:    true,
	}

	c.mu.Lock()
	c.sessions[sessionID] = session
	c.mu.Unlock()

	return session, nil
}

// GetSession retrieves a cached session by ID
func (c *Client) GetSession(sessionID string) (*Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return nil, vonage.ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, vonage.ErrSessionExpired
	}

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one for a spot
func (c *Client) GetOrCreateSession(spotID string, opts *CreateSessionOptions) (*Session, error) {
	// Check cache first
	c.mu.RLock()
	for _, session := range c.sessions {
		if session.SpotID == spotID && session.IsValid() {
			c.mu.RUnlock()
			return session, nil
		}
	}
	c.mu.RUnlock()

	return c.CreateSessionForSpot(spotID, opts)
}

// CleanupExpiredSessions removes expired sessions from the cache
func (c *Client) CleanupExpiredSessions() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for id, session := range c.sessions {
		if session.IsExpired() {
			delete(c.sessions, id)
			count++
		}
	}

	if count > 0 {
		log.Debug().Int("count", count).Msg("Cleaned up expired video sessions")
	}

	return count
}

// CachedSessionCount returns the number of cached sessions
func (c *Client) CachedSessionCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sessions)
}
