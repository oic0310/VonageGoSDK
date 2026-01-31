# Vonage Go SDK

個人的に作っているVonage API Go SDK です。Voiceなどは別途作るかも？

## 構成

```
pkg/vonage/
├── auth.go             # JWT生成・RSA鍵パース
├── client.go           # 統合クライアント & 共通設定
├── errors.go           # カスタムエラー型
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

- [ ] Voice API (`pkg/vonage/voice`)
- [ ] Messages API (`pkg/vonage/messages`)
- [ ] Verify API (`pkg/vonage/verify`)
