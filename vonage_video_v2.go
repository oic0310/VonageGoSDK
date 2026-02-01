package service

import (
	"github.com/rs/zerolog/log"

	"github.com/vonatrigger/poc/internal/config"
	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/video"
)

// VonageVideoServiceV2 wraps the new SDK-based video client
// This provides backward compatibility with the existing service interface
type VonageVideoServiceV2 struct {
	client   *video.Client
	tokenGen *video.TokenGenerator
	appID    string
}

// NewVonageVideoServiceV2 creates a new video service using the SDK
func NewVonageVideoServiceV2(cfg *config.Config, secrets VonageVideoSecrets) (*VonageVideoServiceV2, error) {
	// Create credentials using the SDK
	creds, err := vonage.NewCredentials(
		vonage.WithApplication(secrets.AppID, secrets.PrivateKey),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create Vonage credentials, will use mock mode")
		// Return service in mock mode
		return &VonageVideoServiceV2{
			client:   video.NewClient(secrets.AppID, nil),
			tokenGen: video.NewTokenGenerator(secrets.AppID, nil),
			appID:    secrets.AppID,
		}, nil
	}

	// Create the SDK client
	client, err := video.NewClientFromCredentials(creds)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create Video client, will use mock mode")
		return &VonageVideoServiceV2{
			client:   video.NewClient(secrets.AppID, nil),
			tokenGen: video.NewTokenGenerator(secrets.AppID, nil),
			appID:    secrets.AppID,
		}, nil
	}

	// Create token generator
	jwtGen := vonage.NewJWTGenerator(creds.AppID, creds.PrivateKey)
	tokenGen := video.NewTokenGenerator(creds.AppID, jwtGen)

	log.Info().
		Str("appID", secrets.AppID).
		Msg("Vonage Video API V2 (SDK) configured")

	return &VonageVideoServiceV2{
		client:   client,
		tokenGen: tokenGen,
		appID:    secrets.AppID,
	}, nil
}

// IsConfigured returns true if the service has valid credentials
func (s *VonageVideoServiceV2) IsConfigured() bool {
	return s.client.IsConfigured()
}

// CreateSession creates a new video session via Vonage Video API
// Backward compatible with the old interface
func (s *VonageVideoServiceV2) CreateSession(spotID string) (*VideoSession, error) {
	session, err := s.client.CreateSessionForSpot(spotID, nil)
	if err != nil {
		return nil, err
	}

	// Convert to old format for backward compatibility
	return &VideoSession{
		SessionID: session.SessionID,
		SpotID:    session.SpotID,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// GetSession retrieves an existing session
func (s *VonageVideoServiceV2) GetSession(sessionID string) (*VideoSession, error) {
	session, err := s.client.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return &VideoSession{
		SessionID: session.SessionID,
		SpotID:    session.SpotID,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// GetOrCreateSessionForSpot gets existing session or creates a new one for a spot
func (s *VonageVideoServiceV2) GetOrCreateSessionForSpot(spotID string) (*VideoSession, error) {
	session, err := s.client.GetOrCreateSession(spotID, nil)
	if err != nil {
		return nil, err
	}

	return &VideoSession{
		SessionID: session.SessionID,
		SpotID:    session.SpotID,
		CreatedAt: session.CreatedAt,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// GenerateToken creates a JWT token for a user to join a session
func (s *VonageVideoServiceV2) GenerateToken(sessionID, userID, role string) (*VideoToken, error) {
	var tokenRole video.Role
	switch role {
	case "publisher":
		tokenRole = video.RolePublisher
	case "subscriber":
		tokenRole = video.RoleSubscriber
	case "moderator":
		tokenRole = video.RoleModerator
	default:
		tokenRole = video.RolePublisher
	}

	token, err := s.tokenGen.GenerateToken(sessionID, userID, video.TokenOptions{
		Role: tokenRole,
		Data: userID,
	})
	if err != nil {
		return nil, err
	}

	return &VideoToken{
		Token:     token.Token,
		SessionID: token.SessionID,
		ApiKey:    token.APIKey,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// CleanupExpiredSessions removes expired sessions
func (s *VonageVideoServiceV2) CleanupExpiredSessions() int {
	return s.client.CleanupExpiredSessions()
}

// Client returns the underlying SDK client for advanced usage
func (s *VonageVideoServiceV2) Client() *video.Client {
	return s.client
}

// TokenGenerator returns the underlying token generator for advanced usage
func (s *VonageVideoServiceV2) TokenGenerator() *video.TokenGenerator {
	return s.tokenGen
}
