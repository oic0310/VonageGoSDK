package vonage

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Credentials holds Vonage API credentials
type Credentials struct {
	APIKey      string
	APISecret   string
	AppID       string
	PrivateKey  *rsa.PrivateKey
	PhoneNumber string
}

// CredentialsOption is a functional option for configuring credentials
type CredentialsOption func(*Credentials) error

// WithAPIKey sets the API key and secret
func WithAPIKey(apiKey, apiSecret string) CredentialsOption {
	return func(c *Credentials) error {
		c.APIKey = apiKey
		c.APISecret = apiSecret
		return nil
	}
}

// WithApplication sets the application ID and private key
func WithApplication(appID, privateKeyPEM string) CredentialsOption {
	return func(c *Credentials) error {
		c.AppID = appID
		if privateKeyPEM != "" {
			key, err := ParseRSAPrivateKey(privateKeyPEM)
			if err != nil {
				return fmt.Errorf("failed to parse private key: %w", err)
			}
			c.PrivateKey = key
		}
		return nil
	}
}

// WithPrivateKey sets the private key directly
func WithPrivateKey(key *rsa.PrivateKey) CredentialsOption {
	return func(c *Credentials) error {
		c.PrivateKey = key
		return nil
	}
}

// WithPhoneNumber sets the phone number for outbound calls/SMS
func WithPhoneNumber(number string) CredentialsOption {
	return func(c *Credentials) error {
		c.PhoneNumber = number
		return nil
	}
}

// NewCredentials creates new credentials with the given options
func NewCredentials(opts ...CredentialsOption) (*Credentials, error) {
	c := &Credentials{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// HasApplication returns true if application credentials are configured
func (c *Credentials) HasApplication() bool {
	return c.AppID != "" && c.PrivateKey != nil
}

// HasAPIKey returns true if API key credentials are configured
func (c *Credentials) HasAPIKey() bool {
	return c.APIKey != "" && c.APISecret != ""
}

// ParseRSAPrivateKey parses a PEM encoded RSA private key
func ParseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try PKCS#1 format first
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	// Try PKCS#8 format
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

// JWTGenerator generates JWT tokens for Vonage API authentication
type JWTGenerator struct {
	appID      string
	privateKey *rsa.PrivateKey
}

// NewJWTGenerator creates a new JWT generator
func NewJWTGenerator(appID string, privateKey *rsa.PrivateKey) *JWTGenerator {
	return &JWTGenerator{
		appID:      appID,
		privateKey: privateKey,
	}
}

// JWTClaims represents additional claims for JWT generation
type JWTClaims map[string]interface{}

// GenerateJWT generates a JWT token with the given TTL and additional claims
func (g *JWTGenerator) GenerateJWT(ttl time.Duration, additionalClaims JWTClaims) (string, error) {
	if g.privateKey == nil {
		return "", errors.New("private key not configured")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat":            now.Unix(),
		"exp":            now.Add(ttl).Unix(),
		"jti":            uuid.New().String(),
		"application_id": g.appID,
	}

	// Merge additional claims
	for k, v := range additionalClaims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

// GenerateAPIJWT generates a short-lived JWT for API calls (5 minutes)
func (g *JWTGenerator) GenerateAPIJWT() (string, error) {
	return g.GenerateJWT(5*time.Minute, nil)
}
