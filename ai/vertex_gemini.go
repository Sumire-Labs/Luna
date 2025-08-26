package ai

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/Sumire-Labs/Luna/config"
)

// VertexGeminiService は新しいVertex AI Gemini APIサービス
type VertexGeminiService struct {
	client    *genai.Client
	model     *genai.GenerativeModel
	projectID string
	location  string
}

// NewVertexGeminiService は新しいVertex AI Geminiサービスを作成
func NewVertexGeminiService(cfg *config.GoogleCloudConfig) (*VertexGeminiService, error) {
	ctx := context.Background()

	// Vertex AI クライアントの初期化
	client, err := genai.NewClient(ctx, cfg.ProjectID, cfg.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// モデルの初期化
	model := client.GenerativeModel(cfg.GeminiModel)

	// 安全性設定
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockMediumAndAbove,
		},
	}

	// 生成設定
	model.GenerationConfig = genai.GenerationConfig{
		Temperature:     floatPtr(0.8),
		TopP:            floatPtr(0.95),
		TopK:            intPtr(64),
		MaxOutputTokens: intPtr(2048),
	}

	return &VertexGeminiService{
		client:    client,
		model:     model,
		projectID: cfg.ProjectID,
		location:  cfg.Location,
	}, nil
}

// AskGemini はGeminiに質問して回答を得る
func (s *VertexGeminiService) AskGemini(ctx context.Context, question string, userID string) (string, error) {
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// プロンプトの構築
	prompt := fmt.Sprintf(`あなたは「Luna AI」です。Discord ボット「Luna」に統合された高性能AIアシスタントとして動作しています。

以下のガイドラインに従って回答してください：
- 親切で知識豊富なLuna AIとして振る舞う
- 日本語で丁寧に回答する
- Discord用に最適化された回答（2000文字以内）
- 最新の情報（2025年8月27日時点）に基づいて回答
- 必要に応じて絵文字を使って親しみやすく
- 自分を「Luna AI」または「私」と呼ぶ
- Geminiという名前は一切使わない

ユーザーID: %s
ユーザーの質問: %s`, userID, question)

	// 回答を生成
	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("Gemini APIの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの処理
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Geminiからの応答がありません")
	}

	// 最初の候補から回答を抽出
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("回答の内容が空です")
	}

	// テキストパートを結合
	var result string
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			result += string(textPart)
		}
	}

	if result == "" {
		return "", fmt.Errorf("回答のテキストを取得できませんでした")
	}

	return result, nil
}

// AskGeminiWithImage は画像付きで質問する
func (s *VertexGeminiService) AskGeminiWithImage(ctx context.Context, question string, imageData []byte, mimeType string) (string, error) {
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// プロンプトの構築
	prompt := fmt.Sprintf(`あなたは「Luna AI」です。Discord ボット「Luna」に統合された高性能AIアシスタントです。

提供された画像を分析して、以下の質問に日本語で丁寧に答えてください：
- Luna AIとして親切に回答する
- 画像の詳細な分析結果を提供する
- Discord用に最適化された回答（2000文字以内）
- 必要に応じて絵文字を使用する

質問: %s`, question)

	// 画像とプロンプトで回答を生成
	resp, err := s.model.GenerateContent(ctx,
		genai.Text(prompt),
		genai.ImageData(mimeType, imageData),
	)
	if err != nil {
		return "", fmt.Errorf("Gemini APIの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの処理
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Geminiからの応答がありません")
	}

	// 最初の候補から回答を抽出
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("回答の内容が空です")
	}

	// テキストパートを結合
	var result string
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			result += string(textPart)
		}
	}

	return result, nil
}

// Close はクライアントをクローズする
func (s *VertexGeminiService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Helper functions
func floatPtr(f float32) *float32 {
	return &f
}

func intPtr(i int32) *int32 {
	return &i
}
