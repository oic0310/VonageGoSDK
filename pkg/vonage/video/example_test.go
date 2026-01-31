package video_test

import (
	"fmt"
	"time"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/video"
)

func ExampleClient_basic() {
	// Create credentials
	creds, err := vonage.NewCredentials(
		vonage.WithApplication("your-app-id", `-----BEGIN PRIVATE KEY-----
...your private key...
-----END PRIVATE KEY-----`),
	)
	if err != nil {
		panic(err)
	}

	// Create video client
	client, err := video.NewClientFromCredentials(creds)
	if err != nil {
		panic(err)
	}

	// Create a session
	session, err := client.CreateSession(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Session ID: %s\n", session.SessionID)

	// Generate a token
	tokenGen := video.NewTokenGenerator(creds.AppID, vonage.NewJWTGenerator(creds.AppID, creds.PrivateKey))
	token, err := tokenGen.GeneratePublisherToken(session.SessionID, "user-123")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Token: %s\n", token.Token)
}

func ExampleClient_withOptions() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("your-app-id", "your-private-key-pem"),
	)

	client, _ := video.NewClientFromCredentials(creds)

	// Create session with options
	session, err := client.CreateSession(&video.CreateSessionOptions{
		MediaMode:   video.MediaModeRouted,
		ArchiveMode: video.ArchiveModeManual,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Session: %s\n", session.SessionID)
}

func ExampleTokenGenerator_builder() {
	// Using the fluent token builder
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("your-app-id", "your-private-key-pem"),
	)

	tokenGen := video.NewTokenGenerator(creds.AppID, vonage.NewJWTGenerator(creds.AppID, creds.PrivateKey))

	token, err := tokenGen.NewTokenBuilder("session-id", "user-123").
		WithRole(video.RoleModerator).
		WithTTL(1 * time.Hour).
		WithData(`{"name":"John"}`).
		WithLayoutClasses("full", "focus").
		Build()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Moderator Token: %s\n", token.Token)
}

func ExampleClient_spotManagement() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("your-app-id", "your-private-key-pem"),
	)

	client, _ := video.NewClientFromCredentials(creds)

	// Create or get existing session for a specific spot
	session, err := client.GetOrCreateSession("spot-tokyo-tower", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Session for spot: %s\n", session.SessionID)

	// Cleanup expired sessions periodically
	cleaned := client.CleanupExpiredSessions()
	fmt.Printf("Cleaned up %d expired sessions\n", cleaned)
}
