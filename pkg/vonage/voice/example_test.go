package voice_test

import (
	"context"
	"fmt"

	vonage "github.com/vonatrigger/poc/pkg/vonage"
	"github.com/vonatrigger/poc/pkg/vonage/voice"
)

func ExampleClient_createCall() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := voice.NewClientFromCredentials(creds)

	// Simple call with answer/event URLs
	resp, err := client.CreateCallToPhone(
		context.Background(),
		"81901234567",
		"https://example.com/answer",
		"https://example.com/event",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Call UUID: %s\n", resp.UUID)
}

func ExampleClient_createCallWithNCCO() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
		vonage.WithPhoneNumber("81501234567"),
	)
	client, _ := voice.NewClientFromCredentials(creds)

	// Call with inline NCCO
	ncco := voice.NewNCCO().
		Talk("こんにちは！謎解きイベントへようこそ！").Japanese().Done().
		Build()

	resp, err := client.CreateCallWithNCCO(
		context.Background(),
		"81901234567",
		ncco,
		"https://example.com/event",
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Call UUID: %s\n", resp.UUID)
}

func ExampleClient_callManagement() {
	creds, _ := vonage.NewCredentials(
		vonage.WithApplication("app-id", "private-key-pem"),
	)
	client, _ := voice.NewClientFromCredentials(creds)
	ctx := context.Background()

	callUUID := "some-call-uuid"

	// Get call info
	info, _ := client.GetCallInfo(ctx, callUUID)
	fmt.Printf("Status: %s\n", info.Status)

	// Transfer call to new NCCO
	_ = client.TransferCall(ctx, callUUID, "https://example.com/new-ncco")

	// Mute/unmute
	_ = client.MuteCall(ctx, callUUID)
	_ = client.UnmuteCall(ctx, callUUID)

	// Play TTS into active call
	_ = client.TalkIntoCall(ctx, callUUID, "新しいメッセージです", "Mizuki", 1)

	// Stream audio into active call
	_ = client.StreamIntoCall(ctx, callUUID, "https://example.com/audio.mp3", 1)

	// Hangup
	_ = client.HangupCall(ctx, callUUID)
}

func ExampleNCCOBuilder_basic() {
	// Basic talk NCCO
	ncco := voice.NewNCCO().
		Talk("こんにちは！").Japanese().Done().
		Build()

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleNCCOBuilder_talkAndInput() {
	// Talk then listen for speech (VonaTrigger pattern)
	ncco := voice.NewNCCO().
		Talk("お電話ありがとうございます。何かお手伝いできることはありますか？").
			Japanese().BargeIn().Done().
		Input().Speech().
			EventURL("https://example.com/input").
			EndOnSilence(1.5).
			StartTimeout(5).
			MaxDuration(30).
			Done().
		Build()

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleNCCOBuilder_streamAndInput() {
	// Play pre-synthesized audio then listen (Polly audio pattern)
	ncco := voice.NewNCCO().
		Stream("https://s3.amazonaws.com/bucket/audio.mp3").Done().
		Input().Speech().
			EventURL("https://example.com/input?conversationId=xxx").
			EndOnSilence(1.5).
			StartTimeout(5).
			MaxDuration(30).
			Done().
		Build()

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleNCCOBuilder_dtmfMenu() {
	// DTMF menu with multiple options
	ncco := voice.NewNCCO().
		Talk("メニューを選択してください。1はヒント、2はストーリー、3は終了です。").
			Japanese().BargeIn().Done().
		Input().DTMF().
			EventURL("https://example.com/dtmf-input").
			MaxDigits(1).
			TimeOut(10).
			Done().
		Build()

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleNCCOBuilder_record() {
	// Record a voice message
	ncco := voice.NewNCCO().
		Talk("メッセージを残してください。").Japanese().Done().
		Record().
			Format("mp3").
			BeepStart().
			EndOnSilence(3).
			EndOnKey("#").
			EventURL("https://example.com/recording").
			Done().
		Talk("メッセージを受け付けました。ありがとうございます。").Japanese().Done().
		Build()

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleTalkAndInputJapanese() {
	// Convenience function: Japanese talk + speech input
	ncco := voice.TalkAndInputJapanese(
		"お電話ありがとうございます。謎解きのヒントをお伝えします。",
		"https://example.com/input?conversationId=abc123",
	)

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleStreamAndInput() {
	// Convenience function: Polly audio + speech input
	ncco := voice.StreamAndInput(
		"https://s3.amazonaws.com/bucket/polly-audio.mp3",
		"https://example.com/input?conversationId=abc123",
		1.5,
	)

	data, _ := ncco.JSON()
	fmt.Println(string(data))
}

func ExampleASRResult_processing() {
	// Processing ASR webhook results
	asr := &voice.ASRResult{
		UUID:             "call-uuid",
		ConversationUUID: "conv-uuid",
	}
	asr.Speech.Results = []voice.ASRMatch{
		{Confidence: "0.95", Text: "ヒントをください"},
	}

	if asr.HasSpeech() {
		transcript := asr.BestTranscript()
		fmt.Printf("User said: %s\n", transcript)
	}

	if asr.HasDTMF() {
		fmt.Printf("DTMF: %s\n", asr.DTMF)
	}
}
