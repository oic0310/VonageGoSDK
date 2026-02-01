# VonageGoSDK ドキュメント

Vonage API の Go 言語向け SDK。VonaTrigger プロジェクトのサービス層から抽出・汎用化した、Fluent API ベースの軽量ライブラリです。

**リポジトリ:** [github.com/oic0310/VonageGoSDK](https://github.com/oic0310/VonageGoSDK)

---

## 目次

- [概要](#概要)
- [インストール](#インストール)
- [クイックスタート](#クイックスタート)
- [認証](#認証)
- [Video API](#video-api)
- [Voice API](#voice-api)
- [Messages API](#messages-api)
- [Verify API](#verify-api)
- [移行ガイド](#移行ガイド)
- [アーキテクチャ](#アーキテクチャ)
- [ダイアグラム一覧](#ダイアグラム一覧)

---

## 概要

### 対応 API

| パッケージ | Vonage API | 認証方式 | 主な用途 |
|-----------|------------|---------|---------|
| `pkg/vonage/video` | Video API | JWT | ビデオセッション・トークン生成 |
| `pkg/vonage/voice` | Voice API | JWT | 電話発信・NCCO 制御・通話操作 |
| `pkg/vonage/messages` | Messages API | JWT | SMS / WhatsApp / Viber 送受信 |
| `pkg/vonage/verify` | Verify API | Basic + JWT | 電話番号認証（v1 / v2） |

### パッケージ構成

![パッケージ構成図](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/architecture.puml)

```
VonageGoSDK/
├── pkg/vonage/              # SDK 本体（公開パッケージ）
│   ├── auth.go              #   JWT 生成・RSA 鍵パース
│   ├── client.go            #   統合クライアント
│   ├── errors.go            #   エラー型
│   ├── video/               #   Video API
│   │   ├── client.go
│   │   ├── token.go
│   │   └── types.go
│   ├── voice/               #   Voice API
│   │   ├── client.go
│   │   ├── ncco.go
│   │   └── types.go
│   ├── messages/            #   Messages API
│   │   ├── client.go
│   │   ├── webhook.go
│   │   └── types.go
│   └── verify/              #   Verify API
│       ├── client.go
│       └── types.go
├── internal/service/        # VonaTrigger 向け後方互換ラッパー
│   ├── vonage_video_v2.go
│   ├── vonage_voice_v2.go
│   ├── vonage_messages_v2.go
│   └── vonage_verify_v2.go
├── service/                 # 旧サービス層（参考）
├── go.mod
└── README.md
```

### 設計原則

- **Fluent API** — メソッドチェーンで直感的に構築
- **Functional Options** — `WithXxx()` パターンで柔軟な設定
- **後方互換ラッパー** — 既存コードを壊さずに段階移行
- **ゼロ依存コア** — 標準ライブラリ + JWT + zerolog のみ

---

## インストール

```bash
go get github.com/oic0310/VonageGoSDK
```

### 依存関係

```
github.com/golang-jwt/jwt/v5  # JWT 生成
github.com/google/uuid         # UUID 生成
github.com/rs/zerolog           # 構造化ログ
```

---

## クイックスタート

### SMS を送る（最短コード）

```go
package main

import (
    "context"
    "fmt"

    vonage "github.com/oic0310/VonageGoSDK/pkg/vonage"
    "github.com/oic0310/VonageGoSDK/pkg/vonage/messages"
)

func main() {
    creds, _ := vonage.NewCredentials(
        vonage.WithApplication("your-app-id", yourPrivateKeyPEM),
        vonage.WithPhoneNumber("81501234567"),
    )
    client, _ := messages.NewClientFromCredentials(creds)

    resp, _ := client.SendSMS(context.Background(), "81901234567", "Hello from VonageGoSDK!")
    fmt.Println("Sent:", resp.MessageUUID)
}
```

### 電話をかけて日本語で話す

```go
package main

import (
    "context"
    "fmt"

    vonage "github.com/oic0310/VonageGoSDK/pkg/vonage"
    "github.com/oic0310/VonageGoSDK/pkg/vonage/voice"
)

func main() {
    creds, _ := vonage.NewCredentials(
        vonage.WithApplication("your-app-id", yourPrivateKeyPEM),
        vonage.WithPhoneNumber("81501234567"),
    )
    client, _ := voice.NewClientFromCredentials(creds)

    ncco := voice.TalkJapanese("こんにちは！VonageGoSDK からの発信です。")
    resp, _ := client.CreateCallWithNCCO(
        context.Background(), "81901234567", ncco,
        "https://example.com/event",
    )
    fmt.Println("Call UUID:", resp.UUID)
}
```

---

## 認証

すべての API は `vonage.Credentials` を起点に認証情報を管理します。

### 認証フロー

![認証フロー](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/auth-flow.puml)

### Credentials の作成

```go
import vonage "github.com/oic0310/VonageGoSDK/pkg/vonage"

// JWT 認証（Video / Voice / Messages API 用）
creds, err := vonage.NewCredentials(
    vonage.WithApplication("app-id", privateKeyPEM),
    vonage.WithPhoneNumber("81501234567"),        // 発信元番号（任意）
)

// Basic 認証（Verify API v1 用）
creds, err := vonage.NewCredentials(
    vonage.WithAPIKey("api-key", "api-secret"),
)

// 両方（Verify v1 + v2 対応）
creds, err := vonage.NewCredentials(
    vonage.WithApplication("app-id", privateKeyPEM),
    vonage.WithAPIKey("api-key", "api-secret"),
    vonage.WithPhoneNumber("81501234567"),
)
```

### RSA 秘密鍵の読み込み

```go
// ファイルから読み込み
keyBytes, _ := os.ReadFile("private.key")
privateKeyPEM := string(keyBytes)

// 環境変数から
privateKeyPEM := os.Getenv("VONAGE_PRIVATE_KEY")

// 直接パース
rsaKey, err := vonage.ParseRSAPrivateKey(privateKeyPEM)
creds, _ := vonage.NewCredentials(
    vonage.WithApplication("app-id", ""),
    vonage.WithPrivateKey(rsaKey),
)
```

### JWT の手動生成

```go
jwtGen := vonage.NewJWTGenerator(appID, rsaPrivateKey)

// API 呼び出し用（5 分間有効）
token, err := jwtGen.GenerateAPIJWT()

// カスタム TTL + 追加クレーム
token, err := jwtGen.GenerateJWT(24*time.Hour, vonage.JWTClaims{
    "sub": "user123",
})
```

### Credentials のチェック

```go
creds.HasApplication()  // AppID + 秘密鍵が設定済みか
creds.HasAPIKey()       // APIKey + Secret が設定済みか
```

---

## Video API

ビデオセッションの作成、トークン生成、録画・ブロードキャスト管理を行います。

### セッション管理シーケンス

![Video APIシーケンス](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/video-session-sequence.puml)

### クライアント作成

```go
import "github.com/oic0310/VonageGoSDK/pkg/vonage/video"

client, err := video.NewClientFromCredentials(creds)
```

### セッション作成

```go
// 自動リレーセッション（デフォルト）
session, err := client.CreateSession(ctx, nil)

// P2P セッション
session, err := client.CreateSession(ctx, &video.CreateSessionOptions{
    MediaMode: video.MediaModeRelayed,
})

// 特定地域のセッション
session, err := client.CreateSession(ctx, &video.CreateSessionOptions{
    Location: "ap-northeast-1",
})
```

### トークン生成（Fluent Builder）

```go
// Publisher トークン
token, err := client.GenerateToken(session.SessionID).
    Role(video.RolePublisher).
    ExpireTime(time.Now().Add(2 * time.Hour)).
    Data("userName=田中太郎").
    Build()

// Subscriber トークン（閲覧のみ）
token, err := client.GenerateToken(session.SessionID).
    Role(video.RoleSubscriber).
    Build()

// Moderator トークン
token, err := client.GenerateToken(session.SessionID).
    Role(video.RoleModerator).
    Build()
```

### セッション + トークンの一括作成

```go
// よくあるパターン：セッション作成 → トークン発行
session, err := client.CreateSession(ctx, nil)
token, err := client.GenerateToken(session.SessionID).
    Role(video.RolePublisher).
    Build()

fmt.Println("Session:", session.SessionID)
fmt.Println("Token:", token)
```

### 録画制御

```go
// 録画開始
recording, err := client.StartRecording(ctx, sessionID, nil)

// 録画停止
recording, err := client.StopRecording(ctx, recording.ID)

// 録画一覧
recordings, err := client.ListRecordings(ctx, sessionID)
```

### ブロードキャスト（ライブ配信）

```go
broadcast, err := client.StartBroadcast(ctx, sessionID, &video.BroadcastOptions{
    Outputs: video.BroadcastOutputs{
        HLS: &video.HLSConfig{},
    },
})
```

---

## Voice API

電話の発信、NCCO による通話制御、通話中操作（TTS 注入、ミュート等）、ASR 処理を行います。

### クライアント作成

```go
import "github.com/oic0310/VonageGoSDK/pkg/vonage/voice"

client, err := voice.NewClientFromCredentials(creds)
```

### 電話発信

2 つの発信方式があります。

#### Answer URL 方式

Vonage がユーザー応答時にサーバーの Answer URL を呼び出し、サーバーが NCCO を返すパターンです。

![Answer URL方式](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/voice-answer-url.puml)

```go
resp, err := client.CreateCallToPhone(ctx, "81901234567",
    "https://example.com/answer",
    "https://example.com/event",
)
```

#### インライン NCCO 方式

発信時に NCCO を直接指定するパターンです。Webhook サーバー不要で簡潔です。

![AI通話シーケンス](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/voice-call-sequence.puml)

```go
ncco := voice.TalkJapanese("お電話ありがとうございます。")
resp, err := client.CreateCallWithNCCO(ctx, "81901234567", ncco,
    "https://example.com/event",
)
```

### NCCO Builder（Fluent API）

NCCO（Nexmo Call Control Object）は通話の動作を定義する JSON 配列です。Builder パターンで型安全に構築できます。

![NCCO Builder構造](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/ncco-builder.puml)

#### 基本的な組み立て方

```go
ncco := voice.NewNCCO().
    Talk("お電話ありがとうございます。").Japanese().BargeIn().Done().
    Input().Speech().EventURL("https://example.com/input").
        EndOnSilence(1.5).StartTimeout(5).MaxDuration(30).Done().
    Build()
```

このコードは以下の JSON を生成します：

```json
[
  {
    "action": "talk",
    "text": "お電話ありがとうございます。",
    "voiceName": "Mizuki",
    "language": "ja-JP",
    "bargeIn": true
  },
  {
    "action": "input",
    "type": ["speech"],
    "eventUrl": ["https://example.com/input"],
    "eventMethod": "POST",
    "endOnSilence": 1.5,
    "startTimeout": 5,
    "maxDuration": 30
  }
]
```

#### Talk アクション（TTS）

```go
voice.NewNCCO().
    Talk("Hello!").
        VoiceName("Amy").         // ボイス名
        Language("en-GB").        // 言語
        Style(1).                 // スタイル番号
        Premium().                // プレミアムボイス有効
        Level(0).                 // 音量 (-1 〜 1)
        BargeIn().                // 割り込み許可
        Loop(2).                  // 繰り返し回数
    Done().
    Build()

// 日本語ショートカット（VoiceName=Mizuki, Language=ja-JP を自動設定）
voice.NewNCCO().Talk("こんにちは").Japanese().Done().Build()
```

#### Stream アクション（音声ファイル再生）

```go
voice.NewNCCO().
    Stream("https://example.com/audio.mp3").
        Level(0).
        BargeIn().
        Loop(1).
    Done().
    Build()
```

#### Input アクション（音声認識 / DTMF）

```go
// 音声認識（ASR）
voice.NewNCCO().
    Input().
        Speech().                              // 音声入力を有効化
        EventURL("https://example.com/input"). // 結果送信先
        EndOnSilence(1.5).                     // 1.5 秒無音で終了
        StartTimeout(5).                       // 5 秒以内に話し始める
        MaxDuration(30).                       // 最大 30 秒
    Done().
    Build()

// DTMF（電話のプッシュボタン）
voice.NewNCCO().
    Input().
        DTMF().
        MaxDigits(4).
        SubmitOnHash().  // # キーで送信
        TimeOut(10).
        EventURL("https://example.com/dtmf").
    Done().
    Build()

// 音声 + DTMF 同時受付
voice.NewNCCO().
    Input().SpeechAndDTMF().EventURL(url).Done().
    Build()
```

#### Record アクション（録音）

```go
voice.NewNCCO().
    Record().
        Format("mp3").
        BeepStart().                              // 録音開始ビープ
        EndOnKey("#").                             // # キーで停止
        EndOnSilence(3).                          // 3 秒無音で停止
        EventURL("https://example.com/recording").
        Split().                                  // 通話分割録音
        Channels(2).
    Done().
    Build()
```

#### Notify アクション

```go
voice.NewNCCO().
    Notify("https://example.com/notify", map[string]interface{}{
        "event": "call_started",
        "spotId": "spot-001",
    }).
    Build()
```

#### 複合パターン（実際のユースケース）

```go
// AI 通話パターン：挨拶 → 音声入力待ち → 録音
ncco := voice.NewNCCO().
    Talk("お電話ありがとうございます。ご用件をどうぞ。").Japanese().BargeIn().Done().
    Input().Speech().EventURL(inputURL).
        EndOnSilence(1.5).StartTimeout(5).MaxDuration(30).Done().
    Record().Format("mp3").EventURL(recordURL).EndOnSilence(5).Done().
    Talk("ありがとうございました。失礼いたします。").Japanese().Done().
    Build()
```

### ショートカット関数

よく使うパターンを 1 行で生成できます。

```go
// 日本語 TTS のみ
ncco := voice.TalkJapanese("こんにちは！")

// 日本語 TTS → 音声入力
ncco := voice.TalkAndInputJapanese("ご用件をどうぞ。", inputEventURL)

// 音声ファイル再生 → 音声入力
ncco := voice.StreamAndInput(pollyAudioURL, inputEventURL, 1.5)

// 汎用：TTS → 音声入力（言語・ボイス指定）
ncco := voice.TalkAndInput("Hello!", "en-US", "Amy", inputEventURL, 2.0)
```

### 通話中操作

```go
// 通話情報の取得
info, err := client.GetCallInfo(ctx, callUUID)

// ミュート / ミュート解除
client.MuteCall(ctx, callUUID)
client.UnmuteCall(ctx, callUUID)

// イヤーマフ（相手の音声を聞こえなくする）
client.EarmuffCall(ctx, callUUID)
client.UnearmuffCall(ctx, callUUID)

// 通話中に TTS を注入
client.TalkIntoCall(ctx, callUUID, "新しいヒントです！", "Mizuki", 1)
client.StopTalk(ctx, callUUID)

// 通話中に音声ストリームを注入
client.StreamIntoCall(ctx, callUUID, "https://example.com/hint.mp3", 1)
client.StopStream(ctx, callUUID)

// DTMF 送信
client.SendDTMF(ctx, callUUID, "1234")

// 通話転送（新しい NCCO URL へ）
client.TransferCall(ctx, callUUID, "https://example.com/new-answer")

// 通話終了
client.HangupCall(ctx, callUUID)
```

### ASR 結果の処理

Vonage の音声認識結果を Webhook で受け取った際のヘルパーです。

```go
// Webhook ハンドラ内で
var asr voice.ASRResult
if err := json.NewDecoder(r.Body).Decode(&asr); err != nil {
    // エラー処理
}

// 音声認識結果があるか
if asr.HasSpeech() {
    transcript := asr.BestTranscript() // 最高信頼度のテキスト
    fmt.Println("ユーザー発話:", transcript)
}

// DTMF 入力があるか
if asr.HasDTMF() {
    digits := asr.DTMFDigits
    fmt.Println("DTMF:", digits)
}
```

### 通話イベントの判定

```go
var event voice.CallEvent
json.NewDecoder(r.Body).Decode(&event)

if event.IsTerminal() {
    fmt.Println("通話終了:", event.Status)
}
```

---

## Messages API

SMS / MMS / WhatsApp / Viber などのマルチチャネルメッセージ送受信を統一 API で行います。

### SMS 送受信シーケンス

![Messages APIシーケンス](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/messages-sequence.puml)

### クライアント作成

```go
import "github.com/oic0310/VonageGoSDK/pkg/vonage/messages"

client, err := messages.NewClientFromCredentials(creds)
```

### SMS 送信

```go
// シンプルな SMS
resp, err := client.SendSMS(ctx, "81901234567", "こんにちは！")
fmt.Println("Message UUID:", resp.MessageUUID)

// 送信元番号を指定
resp, err := client.SendSMSFrom(ctx, "81501234567", "81901234567", "ヒントです！")

// オプション付き
resp, err := client.SendSMS(ctx, "81901234567", "追跡用メッセージ",
    messages.WithClientRef("hint-spot-001"),      // ステータス追跡用 ID
    messages.WithWebhookURL("https://example.com/status"), // Webhook URL 上書き
)
```

### Message Builder（Fluent API）

チャネルやコンテンツタイプを柔軟に組み立てられます。

```go
// SMS テキスト
resp, err := client.NewMessage().
    To("81901234567").
    SMS().
    Text("謎解きのヒントです！").
    ClientRef("conv-abc123").
    Send(ctx)

// WhatsApp 画像
resp, err := client.NewMessage().
    To("81901234567").
    WhatsApp().
    Image("https://example.com/clue.jpg", "謎の手がかり").
    Send(ctx)

// Viber テキスト
resp, err := client.NewMessage().
    To("81901234567").
    Viber().
    Text("Viber からのメッセージです！").
    Send(ctx)

// WhatsApp 音声
resp, err := client.NewMessage().
    To("81901234567").
    WhatsApp().
    Audio("https://example.com/voice-hint.mp3").
    Send(ctx)

// ファイル送信
resp, err := client.NewMessage().
    To("81901234567").
    WhatsApp().
    File("https://example.com/report.pdf", "レポート.pdf").
    Send(ctx)
```

### マルチチャネル便利メソッド

```go
// WhatsApp テキスト
client.SendWhatsApp(ctx, "81901234567", "WhatsApp メッセージ")

// WhatsApp 画像
client.SendWhatsAppImage(ctx, "81901234567", "https://example.com/map.jpg", "地図")

// MMS 画像
client.SendMMS(ctx, "81901234567", "https://example.com/clue.jpg", "手がかり")
```

### Webhook ハンドリング

![Webhookハンドリングフロー](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/webhook-handling.puml)

#### net/http 向けハンドラー

```go
handler := messages.NewWebhookHandler().
    OnInbound(func(msg *messages.InboundMessage) error {
        fmt.Printf("受信: %s → %s\n", msg.From, msg.Text)
        fmt.Printf("チャネル: %s, タイプ: %s\n", msg.Channel, msg.MessageType)

        // 画像メッセージの場合
        if msg.Image != nil {
            fmt.Printf("画像 URL: %s\n", msg.Image.URL)
        }
        return nil
    }).
    OnStatus(func(status *messages.MessageStatus) error {
        fmt.Printf("ステータス: %s → %s\n", status.MessageUUID, status.Status)

        if status.Status.IsFailed() {
            fmt.Printf("エラー: %s\n", status.Error.Detail)
        }
        if status.Status.IsDelivered() {
            fmt.Println("配信完了！")
        }
        return nil
    }).
    OnLegacySMS(func(sms *messages.InboundSMS) error {
        // 旧 SMS API フォーマットの処理
        fmt.Printf("旧形式 SMS: %s → %s\n", sms.MSISDN, sms.Text)
        return nil
    })

// HTTP ルーター登録
http.HandleFunc("/webhooks/sms/inbound", handler.HandleInbound())
http.HandleFunc("/webhooks/sms/status", handler.HandleStatus())
```

#### Echo / Gin フレームワーク向けパーサー

```go
// Echo ハンドラ内で
func handleInbound(c echo.Context) error {
    body, _ := io.ReadAll(c.Request().Body)

    // 新旧フォーマット自動判別
    msg, err := messages.ParseInboundMessage(body)
    if err != nil {
        return c.NoContent(http.StatusOK)
    }

    fmt.Printf("From: %s, Text: %s\n", msg.From, msg.Text)
    return c.NoContent(http.StatusOK)
}
```

### ステータス定数と判定

```go
messages.StatusSubmitted  // "submitted" — 送信済み
messages.StatusDelivered  // "delivered" — 配信完了
messages.StatusRead       // "read"      — 既読
messages.StatusRejected   // "rejected"  — 拒否
messages.StatusFailed     // "failed"    — 失敗

status.Status.IsDelivered()  // delivered or read
status.Status.IsFailed()     // rejected or failed
status.Status.IsTerminal()   // 最終状態か（上記いずれか）
```

### チャネル定数

```go
messages.ChannelSMS        // "sms"
messages.ChannelMMS        // "mms"
messages.ChannelWhatsApp   // "whatsapp"
messages.ChannelViber      // "viber_service"
messages.ChannelMessenger  // "messenger"
```

---

## Verify API

電話番号認証を行います。v1（レガシー）と v2（マルチチャネル）の両方に対応し、統一インターフェースで利用できます。

### 電話番号認証シーケンス

![Verify APIシーケンス](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/verify-sequence.puml)

### クライアント作成

```go
import "github.com/oic0310/VonageGoSDK/pkg/vonage/verify"

// v1 のみ（Basic Auth）
creds, _ := vonage.NewCredentials(
    vonage.WithAPIKey("api-key", "api-secret"),
)
client, err := verify.NewClientFromCredentials(creds,
    verify.WithBrand("MyApp"),
    verify.WithLocale("ja-jp"),
    verify.WithCodeLength(6),
)

// v1 + v2 両対応
creds, _ := vonage.NewCredentials(
    vonage.WithAPIKey("api-key", "api-secret"),
    vonage.WithApplication("app-id", privateKeyPEM),
)
client, err := verify.NewClientFromCredentials(creds)
```

### 基本フロー

```go
ctx := context.Background()

// Step 1: 認証開始（SMS でコード送信）
result, err := client.StartVerification(ctx, "81901234567", nil)
if err != nil {
    log.Fatal(err)
}
requestID := result.RequestID  // → DB に保存

// Step 2: ユーザーがコードを入力
userCode := "123456"

// Step 3: コード検証
check, err := client.CheckVerification(ctx, requestID, userCode)
if err != nil {
    log.Fatal(err)
}

if check.Verified {
    fmt.Println("電話番号認証成功！")
    // → ユーザー情報を更新
} else {
    fmt.Printf("認証失敗（ステータス: %s）\n", check.Status)
}
```

### カスタムオプション（v1）

```go
result, err := client.StartVerification(ctx, "81901234567", &verify.StartOptions{
    Brand:      "VonaTrigger",
    CodeLength: 4,                  // 4 桁コード
    Locale:     "ja-jp",            // 日本語 SMS
    PINExpiry:  300,                // 5 分で有効期限切れ
    WorkflowID: verify.WorkflowSMS, // SMS のみ（音声通話なし）
})
```

### Verify v2（マルチチャネルフォールバック）

SMS → WhatsApp → 音声の順に自動フォールバックします。

```go
result, err := client.StartVerification(ctx, "81901234567", &verify.StartOptions{
    Brand: "VonaTrigger",
    V2Channels: []verify.V2Channel{
        verify.V2ChannelSMS,       // まず SMS
        verify.V2ChannelWhatsApp,  // 60 秒後に WhatsApp
        verify.V2ChannelVoice,     // さらに 60 秒後に音声通話
    },
    ChannelTimeout: 60,  // チャネル切り替えまでの秒数
})

// チェックは v1 と同じ
check, err := client.CheckVerification(ctx, result.RequestID, code)
```

### 認証のキャンセル

```go
err := client.CancelVerification(ctx, requestID)
```

### v1 / v2 を明示的に指定

通常は `StartVerification()` がオプションに応じて自動選択しますが、明示指定も可能です。

```go
// 明示的に v1 を使用
result, err := client.StartV1(ctx, "81901234567", opts)
check, err := client.CheckV1(ctx, requestID, code)

// 明示的に v2 を使用
result, err := client.StartV2(ctx, "81901234567", opts)
check, err := client.CheckV2(ctx, requestID, code)
```

### v1 Workflow ID 一覧

| 定数 | ID | フロー |
|------|----|--------|
| `WorkflowSMSTTSTTS` | 1 | SMS → 音声 → 音声（デフォルト） |
| `WorkflowSMSSMSTTS` | 2 | SMS → SMS → 音声 |
| `WorkflowTTSTTSTTS` | 3 | 音声 → 音声 → 音声 |
| `WorkflowSMSSMS` | 4 | SMS → SMS |
| `WorkflowSMSTTS` | 5 | SMS → 音声 |
| `WorkflowSMS` | 6 | SMS のみ |
| `WorkflowTTS` | 7 | 音声のみ |

### v2 チャネル一覧

| 定数 | 説明 |
|------|------|
| `V2ChannelSMS` | SMS |
| `V2ChannelWhatsApp` | WhatsApp |
| `V2ChannelVoice` | 音声通話 |
| `V2ChannelEmail` | メール |
| `V2ChannelSilentAuth` | サイレント認証 |

---

## 移行ガイド

既存の `VonageService` を使ったコードを SDK ベースに段階移行できます。

### 移行パターン

![移行パターン](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/migration.puml)

`internal/service/` の V2 ラッパーが既存インターフェースを実装しているため、呼び出し側の変更は最小限です。

### Voice API の移行

```go
// Before（旧サービス）
vonageService := service.NewVonageService(cfg, secrets)
callResp, err := vonageService.CreateCall(ctx, phone, answerPath, eventPath)
ncco := vonageService.GenerateNCCO(text, inputURL)

// After（V2 ラッパー経由 — インターフェース互換）
vonageService, err := service.NewVonageServiceV2(cfg, secrets)
callResp, err := vonageService.CreateCall(ctx, phone, answerPath, eventPath) // 同じ呼び出し
ncco := vonageService.GenerateNCCO(text, inputURL)                          // 同じ呼び出し

// After（SDK 直接利用 — フル機能）
voiceClient := vonageService.VoiceClient()
ncco := voice.TalkAndInputJapanese(text, inputURL)
voiceClient.CreateCallWithNCCO(ctx, phone, ncco, eventURL)
```

### Messages API の移行

```go
// Before
smsResp, err := vonageService.SendSMS(ctx, phone, text)

// After（V2 ラッパー）
msgService, err := service.NewVonageMessagesServiceV2(cfg, secrets)
smsResp, err := msgService.SendSMS(ctx, phone, text)  // 同じシグネチャ

// After（SDK 直接利用）
msgClient := msgService.Client()
msgClient.SendWhatsApp(ctx, phone, text)  // WhatsApp にも送れる！
```

### Verify API の移行

```go
// Before
verifyResp, err := vonageService.StartVerification(ctx, phone)
checkResp, err := vonageService.CheckVerification(ctx, requestID, code)

// After（V2 ラッパー）
verifyService, err := service.NewVonageVerifyServiceV2(cfg, secrets)
verifyResp, err := verifyService.StartVerification(ctx, phone)   // 同じ
checkResp, err := verifyService.CheckVerification(ctx, reqID, code) // 同じ

// After（SDK 直接利用 — v2 マルチチャネル対応）
verifyClient := verifyService.Client()
result, err := verifyClient.StartVerification(ctx, phone, &verify.StartOptions{
    V2Channels: []verify.V2Channel{verify.V2ChannelSMS, verify.V2ChannelWhatsApp},
})
```

---

## アーキテクチャ

### VonaTrigger 統合フロー

SDK の全 API を組み合わせた VonaTrigger の謎解きイベントフローです。

![VonaTrigger統合フロー](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/vonatrigger-integration.puml)

### エラーハンドリング

```go
import vonage "github.com/oic0310/VonageGoSDK/pkg/vonage"

resp, err := client.SendSMS(ctx, to, text)
if err != nil {
    // Vonage API エラーかチェック
    var apiErr *vonage.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Body)
    }
    // その他のエラー（ネットワーク、JSON パース等）
    return err
}
```

### 定義済みエラー

```go
vonage.ErrNotConfigured  // 認証情報が未設定
vonage.ErrUnauthorized   // 認証失敗（401）
vonage.ErrForbidden      // アクセス拒否（403）
vonage.ErrNotFound       // リソースが見つからない（404）
vonage.ErrServerError    // サーバーエラー（5xx）
```

---

## ダイアグラム一覧

すべてのダイアグラムは `doc/diagrams/` ディレクトリに PlantUML 形式で格納されています。

| ファイル | 種類 | 内容 |
|---------|------|------|
| [`architecture.puml`](diagrams/architecture.puml) | コンポーネント図 | パッケージ構成と依存関係 |
| [`auth-flow.puml`](diagrams/auth-flow.puml) | シーケンス図 | JWT / Basic 認証フロー |
| [`voice-call-sequence.puml`](diagrams/voice-call-sequence.puml) | シーケンス図 | AI 通話（インライン NCCO 方式） |
| [`voice-answer-url.puml`](diagrams/voice-answer-url.puml) | シーケンス図 | Answer URL 方式の発信フロー |
| [`ncco-builder.puml`](diagrams/ncco-builder.puml) | アクティビティ図 | NCCO Builder のメソッドチェーン構造 |
| [`messages-sequence.puml`](diagrams/messages-sequence.puml) | シーケンス図 | SMS / マルチチャネル送受信 |
| [`webhook-handling.puml`](diagrams/webhook-handling.puml) | アクティビティ図 | Webhook ハンドリングフロー |
| [`verify-sequence.puml`](diagrams/verify-sequence.puml) | シーケンス図 | 電話番号認証（v1 / v2） |
| [`video-session-sequence.puml`](diagrams/video-session-sequence.puml) | シーケンス図 | セッション管理と録画 |
| [`migration.puml`](diagrams/migration.puml) | コンポーネント図 | 旧サービスからの段階的移行パターン |
| [`vonatrigger-integration.puml`](diagrams/vonatrigger-integration.puml) | シーケンス図 | 全 API 統合利用（イベントフロー） |

### ローカルでの描画

```bash
# PlantUML をインストール
brew install plantuml

# 全図を一括生成
plantuml doc/diagrams/*.puml -o ../images

# 個別生成（SVG）
plantuml doc/diagrams/architecture.puml -tsvg
```

### GitHub での表示

GitHub は PlantUML を直接レンダリングしません。本ドキュメントでは **方法 1（PlantUML Proxy）** を適用済みです。

**方法 1: PlantUML Proxy を利用（✅ 本ドキュメントで採用）**

画像リンクを以下の形式にすると、GitHub 上で自動描画されます：

```markdown
![図の名前](https://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/oic0310/VonageGoSDK/develop/doc/diagrams/architecture.puml)
```

> ⚠️ `.puml` ファイルを更新した場合、PlantUML Proxy のキャッシュにより反映まで数分かかることがあります。
> `&cache=no` パラメータを付与するとキャッシュを回避できます。

**方法 2: 画像を事前生成してコミット**

```bash
plantuml doc/diagrams/*.puml -o ../images -tpng
# doc/images/ に PNG が生成される → git add してコミット
```

**方法 3: GitHub Actions で自動生成**

```yaml
# .github/workflows/plantuml.yml
name: Generate PlantUML diagrams
on:
  push:
    paths:
      - 'doc/diagrams/**'
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: grassedge/generate-plantuml-action@v1
        with:
          path: doc/diagrams
          output: doc/images
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 全 API メソッドリファレンス

| API | メソッド | 説明 |
|-----|---------|------|
| **Video** | `CreateSession()` | セッション作成 |
| | `GenerateToken().Build()` | トークン生成（Fluent） |
| | `StartRecording()` / `StopRecording()` | 録画制御 |
| | `StartBroadcast()` / `StopBroadcast()` | ブロードキャスト |
| **Voice** | `CreateCallToPhone()` | 電話発信（Answer URL） |
| | `CreateCallWithNCCO()` | 電話発信（インライン NCCO） |
| | `NewNCCO().Talk().Done().Build()` | NCCO Builder |
| | `TalkJapanese()` / `TalkAndInputJapanese()` | ショートカット |
| | `MuteCall()` / `TalkIntoCall()` | 通話中操作 |
| | `TransferCall()` / `HangupCall()` | 通話制御 |
| **Messages** | `SendSMS()` | SMS 送信 |
| | `SendWhatsApp()` / `SendWhatsAppImage()` | WhatsApp 送信 |
| | `SendMMS()` | MMS 送信 |
| | `NewMessage().To().SMS().Text().Send()` | Fluent Builder |
| | `NewWebhookHandler().OnInbound().OnStatus()` | Webhook ハンドラ |
| | `ParseInboundMessage()` | 受信パース（新旧自動判別） |
| **Verify** | `StartVerification()` | 認証開始（v1/v2 自動選択） |
| | `CheckVerification()` | コード検証 |
| | `CancelVerification()` | 認証キャンセル |
| | `StartV1()` / `StartV2()` | API バージョン明示指定 |

---


