# Vonage Go SDK

VonaTrigger プロジェクト用の Vonage API Go SDK です。

## 構成

```
pkg/vonage/
├── auth.go             # JWT生成・RSA鍵パース
├── client.go           # 統合クライアント & 共通設定
├── errors.go           # カスタムエラー型
├── voice/
│   ├── client.go       # Voice API クライアント（Call管理）
│   ├── ncco.go         # NCCOビルダー（Fluent API）
│   ├── types.go        # Call, ASRResult, CallEvent等
│   └── example_test.go # 使用例
├── messages/
│   ├── client.go       # Messages API クライアント（SMS/WhatsApp/Viber）
│   ├── webhook.go      # Webhookハンドラー & パーサー
│   ├── types.go        # SendRequest, InboundMessage, Status等
│   └── example_test.go # 使用例
└── video/
    ├── client.go       # Video API クライアント
    ├── token.go        # トークン生成（Fluent Builder対応）
    ├── types.go        # 型定義（Session, Token, Role等）
    └── example_test.go # 使用例
```

## インストール

```bash
go get github.com/vonatrigger/poc/pkg/vonage
```

## 基本的な使い方

### 1. 資格情報の作成

```go
import (
    vonage "github.com/vonatrigger/poc/pkg/vonage"
    "github.com/vonatrigger/poc/pkg/vonage/video"
)

// アプリケーションIDと秘密鍵で認証
creds, err := vonage.NewCredentials(
    vonage.WithApplication("your-app-id", privateKeyPEM),
)
```

### 2. Video APIの使用

```go
// クライアント作成
client, err := video.NewClientFromCredentials(creds)
if err != nil {
    log.Fatal(err)
}

// セッション作成
session, err := client.CreateSession(nil)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Session ID: %s\n", session.SessionID)

// トークン生成
tokenGen := video.NewTokenGenerator(creds.AppID, vonage.NewJWTGenerator(creds.AppID, creds.PrivateKey))
token, err := tokenGen.GeneratePublisherToken(session.SessionID, "user-123")
```

### 3. Fluent Token Builder

```go
token, err := tokenGen.NewTokenBuilder(session.SessionID, "user-123").
    WithRole(video.RoleModerator).
    WithTTL(1 * time.Hour).
    WithData(`{"name":"John"}`).
    WithLayoutClasses("full", "focus").
    Build()
```

### 4. Spot管理（VonaTrigger用）

```go
// 特定のSpotに紐づくセッションを取得または作成
session, err := client.GetOrCreateSession("spot-tokyo-tower", nil)

// 期限切れセッションのクリーンアップ
cleaned := client.CleanupExpiredSessions()
```

## 既存サービスからの移行

既存の `VonageVideoService` から移行する場合は、`VonageVideoServiceV2` を使用してください：

```go
// internal/service/vonage_video_v2.go を使用
videoService, err := service.NewVonageVideoServiceV2(cfg, secrets)

// 同じインターフェースで使用可能
session, err := videoService.CreateSession("spot-id")
token, err := videoService.GenerateToken(session.SessionID, "user-id", "publisher")
```

## セッションオプション

```go
session, err := client.CreateSession(&video.CreateSessionOptions{
    // メディアモード: relayed（P2P）または routed（サーバー経由）
    MediaMode:   video.MediaModeRouted,
    
    // アーカイブモード: manual または always
    ArchiveMode: video.ArchiveModeManual,
    
    // ロケーションヒント（オプション）
    Location:    "Tokyo",
})
```

## ロール

| Role | 説明 |
|------|------|
| `RolePublisher` | ストリームの公開と購読が可能（デフォルト） |
| `RoleSubscriber` | ストリームの購読のみ可能 |
| `RoleModerator` | セッションの完全な制御が可能 |

## エラーハンドリング

```go
session, err := client.GetSession("invalid-id")
if err != nil {
    if err == vonage.ErrSessionNotFound {
        // セッションが見つからない
    } else if err == vonage.ErrSessionExpired {
        // セッションが期限切れ
    } else if apiErr, ok := err.(*vonage.Error); ok {
        // Vonage APIエラー
        if apiErr.IsRateLimited() {
            // レート制限
        }
    }
}
```

## Mock モード

資格情報が設定されていない場合、SDKは自動的にMockモードで動作します。
Mockセッション/トークンは `mock_` プレフィックスで識別できます。

```go
if session.IsMock {
    // Mockセッション - フロントエンドで事前録画ビデオにフォールバック
}
```

## 今後の拡張予定

- [x] Video API (`pkg/vonage/video`)
- [x] Voice API (`pkg/vonage/voice`)
- [x] Messages API (`pkg/vonage/messages`)
- [ ] Verify API (`pkg/vonage/verify`)

---

## Messages APIの使用

### 1. クライアント作成

```go
import (
    vonage "github.com/vonatrigger/poc/pkg/vonage"
    "github.com/vonatrigger/poc/pkg/vonage/messages"
)

creds, err := vonage.NewCredentials(
    vonage.WithApplication("your-app-id", privateKeyPEM),
    vonage.WithPhoneNumber("81501234567"),
)

client, err := messages.NewClientFromCredentials(creds)
```

### 2. SMS送信

```go
// シンプルなSMS送信
resp, err := client.SendSMS(ctx, "81901234567", "こんにちは！謎解きイベントへようこそ！")

// クライアントリファレンス付きSMS（ステータス追跡用）
resp, err := client.SendSMS(ctx, "81901234567", "ヒントです！",
    messages.WithClientRef("hint-spot-001"),
)
```

### 3. Fluent Message Builder

```go
// SMS
resp, err := client.NewMessage().
    To("81901234567").
    SMS().
    Text("謎解きのヒントです！").
    ClientRef("conv-abc123").
    Send(ctx)

// WhatsApp画像
resp, err := client.NewMessage().
    To("81901234567").
    WhatsApp().
    Image("https://example.com/clue.jpg", "謎の手がかり").
    Send(ctx)

// Viber
resp, err := client.NewMessage().
    To("81901234567").
    Viber().
    Text("Viberからのメッセージです！").
    Send(ctx)
```

### 4. マルチチャネル対応

```go
// WhatsApp テキスト
client.SendWhatsApp(ctx, "81901234567", "WhatsAppメッセージ")

// WhatsApp 画像
client.SendWhatsAppImage(ctx, "81901234567", "https://example.com/map.jpg", "ヒント地図")

// MMS 画像
client.SendMMS(ctx, "81901234567", "https://example.com/clue.jpg", "手がかり")
```

### 5. Webhookハンドリング

```go
// net/http用ハンドラー
handler := messages.NewWebhookHandler().
    OnInbound(func(msg *messages.InboundMessage) error {
        fmt.Printf("Received from %s: %s\n", msg.From, msg.Text)
        return nil
    }).
    OnStatus(func(status *messages.MessageStatus) error {
        if status.Status.IsFailed() {
            log.Error().Str("error", status.Error.Detail).Msg("Message failed")
        }
        return nil
    }).
    OnLegacySMS(func(sms *messages.InboundSMS) error {
        // 旧SMS APIフォーマットの処理
        return nil
    })

http.HandleFunc("/webhooks/sms/inbound", handler.HandleInbound())
http.HandleFunc("/webhooks/sms/status", handler.HandleStatus())
```

### 6. Echo/Ginフレームワークでのパース

```go
// Echo handler内で
body, _ := io.ReadAll(c.Request().Body)
msg, err := messages.ParseInboundMessage(body)  // 新旧フォーマット自動判別
status, err := messages.ParseMessageStatus(body)
```

### 7. ステータス判定

```go
if status.Status.IsDelivered() { /* 配信成功 */ }
if status.Status.IsFailed()    { /* 配信失敗 */ }
if status.Status.IsTerminal()  { /* 最終状態（成功 or 失敗）*/ }
```

### 8. 既存サービスからの移行

```go
// internal/service/vonage_messages_v2.go を使用
msgService, err := service.NewVonageMessagesServiceV2(cfg, secrets)

// 同じインターフェースで使用可能
smsResp, err := msgService.SendSMS(ctx, phoneNumber, text)
```

---

## Voice APIの使用

### 1. クライアント作成

```go
import (
    vonage "github.com/vonatrigger/poc/pkg/vonage"
    "github.com/vonatrigger/poc/pkg/vonage/voice"
)

creds, err := vonage.NewCredentials(
    vonage.WithApplication("your-app-id", privateKeyPEM),
    vonage.WithPhoneNumber("81501234567"),
)

client, err := voice.NewClientFromCredentials(creds)
```

### 2. 発信

```go
// Answer URLを使った発信
resp, err := client.CreateCallToPhone(ctx, "81901234567",
    "https://example.com/answer",
    "https://example.com/event",
)

// インラインNCCOを使った発信
ncco := voice.TalkJapanese("こんにちは！謎解きイベントへようこそ！")
resp, err := client.CreateCallWithNCCO(ctx, "81901234567", ncco, "https://example.com/event")
```

### 3. NCCOビルダー（Fluent API）

```go
// 基本パターン: 日本語で話す → 音声入力を待つ
ncco := voice.NewNCCO().
    Talk("お電話ありがとうございます。何かお手伝いできることはありますか？").
        Japanese().BargeIn().Done().
    Input().Speech().
        EventURL("https://example.com/input?conversationId=xxx").
        EndOnSilence(1.5).StartTimeout(5).MaxDuration(30).Done().
    Build()

// Polly音声 → 音声入力パターン
ncco := voice.NewNCCO().
    Stream("https://s3.amazonaws.com/bucket/polly-audio.mp3").Done().
    Input().Speech().
        EventURL("https://example.com/input").
        EndOnSilence(1.5).StartTimeout(5).MaxDuration(30).Done().
    Build()

// DTMFメニュー
ncco := voice.NewNCCO().
    Talk("1はヒント、2はストーリー、3は終了です。").Japanese().BargeIn().Done().
    Input().DTMF().EventURL("https://example.com/dtmf").MaxDigits(1).TimeOut(10).Done().
    Build()

// 録音
ncco := voice.NewNCCO().
    Talk("メッセージを残してください。").Japanese().Done().
    Record().Format("mp3").BeepStart().EndOnSilence(3).
        EventURL("https://example.com/recording").Done().
    Build()
```

### 4. ショートカット関数

```go
// 日本語 Talk + Speech Input (VonaTrigger標準パターン)
ncco := voice.TalkAndInputJapanese(text, inputEventURL)

// Stream + Speech Input (Polly音声パターン)
ncco := voice.StreamAndInput(audioURL, inputEventURL, 1.5)

// Talk のみ (ハングアップ前)
ncco := voice.TalkJapanese(text)
```

### 5. 通話中の操作

```go
// 通話情報の取得
info, _ := client.GetCallInfo(ctx, callUUID)
fmt.Printf("Status: %s, Duration: %s\n", info.Status, info.Duration)

// 通話の転送
client.TransferCall(ctx, callUUID, "https://example.com/new-ncco")

// ミュート / アンミュート
client.MuteCall(ctx, callUUID)
client.UnmuteCall(ctx, callUUID)

// 通話中にTTSを流す
client.TalkIntoCall(ctx, callUUID, "新しいヒントです！", "Mizuki", 1)

// 通話中にストリーム音声を流す
client.StreamIntoCall(ctx, callUUID, "https://example.com/audio.mp3", 1)

// 通話終了
client.HangupCall(ctx, callUUID)
```

### 6. ASR（音声認識）結果の処理

```go
// Webhookハンドラー内で
var asr voice.ASRResult
if err := c.Bind(&asr); err != nil { ... }

if asr.HasSpeech() {
    transcript := asr.BestTranscript()
    // AI応答を生成...
}

if asr.HasDTMF() {
    switch asr.DTMF {
    case "1": // ヒント
    case "2": // ストーリー
    }
}
```

### 7. 既存サービスからの移行

```go
// internal/service/vonage_voice_v2.go を使用
voiceService, err := service.NewVonageServiceV2(cfg, secrets)

// 同じインターフェースで使用可能（後方互換）
callResp, err := voiceService.CreateCall(ctx, phoneNumber, answerPath, eventPath)
ncco := voiceService.GenerateNCCO(text, inputEventURL)
ncco := voiceService.GenerateNCCOWithStream(audioURL, inputEventURL)
```
