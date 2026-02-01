package video

import "time"

// Session represents a Vonage Video session
type Session struct {
	SessionID string    `json:"sessionId"`
	SpotID    string    `json:"spotId,omitempty"`
	ProjectID string    `json:"projectId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	IsMock    bool      `json:"isMock,omitempty"`
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid returns true if the session is valid and not expired
func (s *Session) IsValid() bool {
	return s.SessionID != "" && !s.IsExpired()
}

// Token represents a video session token
type Token struct {
	Token     string `json:"token"`
	SessionID string `json:"sessionId"`
	APIKey    string `json:"apiKey"`
	ExpiresAt int64  `json:"expiresAt"`
}

// Role represents the role of a participant in a video session
type Role string

const (
	// RolePublisher can publish and subscribe to streams
	RolePublisher Role = "publisher"
	// RoleSubscriber can only subscribe to streams
	RoleSubscriber Role = "subscriber"
	// RoleModerator has full control over the session
	RoleModerator Role = "moderator"
)

// MediaMode represents the media mode for a session
type MediaMode string

const (
	// MediaModeRelayed uses peer-to-peer connections
	MediaModeRelayed MediaMode = "relayed"
	// MediaModeRouted uses Vonage's media servers
	MediaModeRouted MediaMode = "routed"
)

// ArchiveMode represents the archive mode for a session
type ArchiveMode string

const (
	// ArchiveModeManual requires manual archive start
	ArchiveModeManual ArchiveMode = "manual"
	// ArchiveModeAlways automatically archives all streams
	ArchiveModeAlways ArchiveMode = "always"
)

// CreateSessionOptions contains options for creating a session
type CreateSessionOptions struct {
	// Location is a preferred location for the session (IP address or geographic location)
	Location string
	// MediaMode determines how streams are routed
	MediaMode MediaMode
	// ArchiveMode determines how streams are archived
	ArchiveMode ArchiveMode
	// P2PPreference is deprecated, use MediaMode instead
	P2PPreference string
}

// CreateSessionResponse represents the Vonage API response for session creation
type CreateSessionResponse struct {
	SessionID      string `json:"session_id"`
	ProjectID      string `json:"project_id"`
	CreateDt       string `json:"create_dt"`
	MediaServerURL string `json:"media_server_url"`
}

// TokenOptions contains options for generating a token
type TokenOptions struct {
	// Role determines the participant's capabilities
	Role Role
	// ExpireTime is when the token expires (default: 24 hours)
	ExpireTime time.Time
	// Data is custom data to include in the token (max 1000 chars)
	Data string
	// InitialLayoutClassList is a list of layout classes for the stream
	InitialLayoutClassList []string
}

// DefaultTokenOptions returns the default token options
func DefaultTokenOptions() TokenOptions {
	return TokenOptions{
		Role:       RolePublisher,
		ExpireTime: time.Now().Add(24 * time.Hour),
	}
}
