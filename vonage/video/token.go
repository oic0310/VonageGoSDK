package video

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
)

// TokenGenerator generates tokens for Vonage Video sessions
type TokenGenerator struct {
	appID        string
	jwtGenerator *vonage.JWTGenerator
}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator(appID string, jwtGenerator *vonage.JWTGenerator) *TokenGenerator {
	return &TokenGenerator{
		appID:        appID,
		jwtGenerator: jwtGenerator,
	}
}

// GenerateToken creates a JWT token for a user to join a video session
func (g *TokenGenerator) GenerateToken(sessionID, userID string, opts TokenOptions) (*Token, error) {
	if g.jwtGenerator == nil {
		return g.generateMockToken(sessionID, userID, opts)
	}

	// Set defaults
	if opts.Role == "" {
		opts.Role = RolePublisher
	}
	if opts.ExpireTime.IsZero() {
		opts.ExpireTime = time.Now().Add(24 * time.Hour)
	}

	now := time.Now()

	// Build claims for Vonage Video client token
	claims := vonage.JWTClaims{
		"iat":        now.Unix(),
		"exp":        opts.ExpireTime.Unix(),
		"jti":        uuid.New().String(),
		"scope":      "session.connect",
		"session_id": sessionID,
		"role":       string(opts.Role),
	}

	// Add optional data
	if opts.Data != "" {
		claims["data"] = opts.Data
	}

	if len(opts.InitialLayoutClassList) > 0 {
		claims["initial_layout_class_list"] = opts.InitialLayoutClassList
	}

	token, err := g.jwtGenerator.GenerateJWT(opts.ExpireTime.Sub(now), claims)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("userID", userID).
		Str("role", string(opts.Role)).
		Msg("Generated Vonage Video token")

	return &Token{
		Token:     token,
		SessionID: sessionID,
		APIKey:    g.appID,
		ExpiresAt: opts.ExpireTime.Unix(),
	}, nil
}

// GeneratePublisherToken is a convenience method to generate a publisher token
func (g *TokenGenerator) GeneratePublisherToken(sessionID, userID string) (*Token, error) {
	return g.GenerateToken(sessionID, userID, TokenOptions{
		Role: RolePublisher,
		Data: userID,
	})
}

// GenerateSubscriberToken is a convenience method to generate a subscriber token
func (g *TokenGenerator) GenerateSubscriberToken(sessionID, userID string) (*Token, error) {
	return g.GenerateToken(sessionID, userID, TokenOptions{
		Role: RoleSubscriber,
		Data: userID,
	})
}

// GenerateModeratorToken is a convenience method to generate a moderator token
func (g *TokenGenerator) GenerateModeratorToken(sessionID, userID string) (*Token, error) {
	return g.GenerateToken(sessionID, userID, TokenOptions{
		Role: RoleModerator,
		Data: userID,
	})
}

// generateMockToken creates a mock token for development/testing
func (g *TokenGenerator) generateMockToken(sessionID, userID string, opts TokenOptions) (*Token, error) {
	if opts.ExpireTime.IsZero() {
		opts.ExpireTime = time.Now().Add(24 * time.Hour)
	}
	if opts.Role == "" {
		opts.Role = RolePublisher
	}

	mockData := map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"role":       string(opts.Role),
		"exp":        opts.ExpireTime.Unix(),
		"mock":       true,
	}

	jsonData, _ := json.Marshal(mockData)
	mockToken := "mock_" + base64.StdEncoding.EncodeToString(jsonData)

	log.Debug().
		Str("sessionID", sessionID).
		Str("userID", userID).
		Msg("Generated mock video token")

	return &Token{
		Token:     mockToken,
		SessionID: sessionID,
		APIKey:    "mock_api_key",
		ExpiresAt: opts.ExpireTime.Unix(),
	}, nil
}

// TokenBuilder provides a fluent API for building token options
type TokenBuilder struct {
	sessionID string
	userID    string
	opts      TokenOptions
	generator *TokenGenerator
}

// NewTokenBuilder creates a new token builder
func (g *TokenGenerator) NewTokenBuilder(sessionID, userID string) *TokenBuilder {
	return &TokenBuilder{
		sessionID: sessionID,
		userID:    userID,
		opts:      DefaultTokenOptions(),
		generator: g,
	}
}

// WithRole sets the role
func (b *TokenBuilder) WithRole(role Role) *TokenBuilder {
	b.opts.Role = role
	return b
}

// WithExpireTime sets the expiration time
func (b *TokenBuilder) WithExpireTime(t time.Time) *TokenBuilder {
	b.opts.ExpireTime = t
	return b
}

// WithTTL sets the expiration time based on a duration from now
func (b *TokenBuilder) WithTTL(ttl time.Duration) *TokenBuilder {
	b.opts.ExpireTime = time.Now().Add(ttl)
	return b
}

// WithData sets custom data
func (b *TokenBuilder) WithData(data string) *TokenBuilder {
	b.opts.Data = data
	return b
}

// WithLayoutClasses sets the initial layout class list
func (b *TokenBuilder) WithLayoutClasses(classes ...string) *TokenBuilder {
	b.opts.InitialLayoutClassList = classes
	return b
}

// Build generates the token
func (b *TokenBuilder) Build() (*Token, error) {
	return b.generator.GenerateToken(b.sessionID, b.userID, b.opts)
}

// ExtendedTokenClaims provides access to custom JWT claims for advanced use cases
type ExtendedTokenClaims struct {
	jwt.RegisteredClaims
	ApplicationID          string   `json:"application_id"`
	Scope                  string   `json:"scope"`
	SessionID              string   `json:"session_id"`
	Role                   string   `json:"role"`
	Data                   string   `json:"data,omitempty"`
	InitialLayoutClassList []string `json:"initial_layout_class_list,omitempty"`
}
