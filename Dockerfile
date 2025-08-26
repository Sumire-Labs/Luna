# 🐳 Luna Bot Dockerfile

# ビルドステージ
FROM golang:1.24.4-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 必要なパッケージをインストール
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Go modules をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty)" \
    -o luna cmd/bot/main.go

# 実行ステージ
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk add --no-cache ca-certificates tzdata

# 実行用ユーザーを作成
RUN addgroup -g 1001 -S luna && \
    adduser -u 1001 -S luna -G luna

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/luna ./
COPY --from=builder /app/config.toml.example ./

# データディレクトリを作成
RUN mkdir -p data && \
    chown -R luna:luna /app

# 実行ユーザーに切り替え
USER luna

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD pgrep luna || exit 1

# ポート（必要に応じて）
# EXPOSE 8080

# ボリューム
VOLUME ["/app/data", "/app/config"]

# エントリーポイント
ENTRYPOINT ["./luna"]