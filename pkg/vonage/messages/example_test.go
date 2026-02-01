package messages_test

import (
	"context"
	"fmt"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/messages"
)

func ExampleClient_sendSMS() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// Simple SMS send
	resp, err := client.SendSMS(context.Background(), "81901234567", "こんにちは！謎解きイベントへようこそ！")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleClient_sendSMSWithOptions() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// SMS with client reference for tracking
	resp, err := client.SendSMS(
		context.Background(),
		"81901234567",
		"ヒント: 東京タワーの近くを探してみてください。",
		messages.WithClientRef("hint-spot-001"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleClient_messageBuilder() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// Fluent API for building messages
	resp, err := client.NewMessage().
		To("81901234567").
		SMS().
		Text("謎解きのヒントです！次のスポットに向かってください。").
		ClientRef("conv-abc123").
		Send(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleClient_sendWhatsApp() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// WhatsApp text message
	resp, err := client.SendWhatsApp(
		context.Background(),
		"81901234567",
		"WhatsAppからのメッセージです！",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleClient_sendWhatsAppImage() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// WhatsApp image with caption
	resp, err := client.SendWhatsAppImage(
		context.Background(),
		"81901234567",
		"https://example.com/map-hint.jpg",
		"次のスポットのヒント地図です！",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleClient_messageBuilderMultiChannel() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := messages.NewClientFromCredentials(creds)

	// WhatsApp image via builder
	resp, err := client.NewMessage().
		To("81901234567").
		WhatsApp().
		Image("https://example.com/clue.jpg", "謎の手がかり").
		Send(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)

	// Viber text via builder
	resp, err = client.NewMessage().
		To("81901234567").
		Viber().
		Text("Viberからのメッセージです！").
		Send(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message UUID: %s\n", resp.MessageUUID)
}

func ExampleWebhookHandler() {
	// Webhook handler setup (works with net/http)
	handler := messages.NewWebhookHandler().
		OnInbound(func(msg *messages.InboundMessage) error {
			fmt.Printf("Received from %s: %s\n", msg.From, msg.Text)
			// Process message, e.g., send to AI for response
			return nil
		}).
		OnStatus(func(status *messages.MessageStatus) error {
			fmt.Printf("Message %s status: %s\n", status.MessageUUID, status.Status)
			if status.Status.IsFailed() {
				fmt.Printf("Message failed: %s\n", status.Error.Detail)
			}
			return nil
		}).
		OnLegacySMS(func(sms *messages.InboundSMS) error {
			fmt.Printf("Legacy SMS from %s: %s\n", sms.MSISDN, sms.Text)
			return nil
		})

	// Register with your HTTP router
	// http.HandleFunc("/webhooks/vonage/sms/inbound", handler.HandleInbound())
	// http.HandleFunc("/webhooks/vonage/sms/status", handler.HandleStatus())
	_ = handler
}

func ExampleParseInboundMessage() {
	// For use with Echo/Gin frameworks
	// In an Echo handler:
	// body, _ := io.ReadAll(c.Request().Body)
	// msg, err := messages.ParseInboundMessage(body)

	// Legacy SMS format
	legacyBody := []byte(`{"msisdn":"81901234567","to":"81501234567","messageId":"msg-001","text":"ヒントください"}`)
	msg, err := messages.ParseInboundMessage(legacyBody)
	if err != nil {
		panic(err)
	}
	fmt.Printf("From: %s, Text: %s, Channel: %s\n", msg.From, msg.Text, msg.Channel)

	// Messages API format
	newBody := []byte(`{"message_uuid":"uuid-001","from":"81901234567","to":"81501234567","channel":"sms","message_type":"text","text":"答えは42です"}`)
	msg, err = messages.ParseInboundMessage(newBody)
	if err != nil {
		panic(err)
	}
	fmt.Printf("UUID: %s, Text: %s\n", msg.MessageUUID, msg.Text)
}
