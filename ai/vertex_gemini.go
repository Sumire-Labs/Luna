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

	// 候補の存在を確認
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("AIからの回答候補がありません")
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

	// 候補の存在を確認
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("AIからの回答候補がありません")
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

// OCRWithGemini は画像からテキストを抽出する（Vertex AI版）
func (s *VertexGeminiService) OCRWithGemini(ctx context.Context, imageData []byte, mimeType string, userID string, mode string) (string, error) {
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	// モードに応じたプロンプトを構築
	var prompt string
	switch mode {
	case "text":
		prompt = `この画像に含まれるすべてのテキストを正確に抽出してください。
以下の点に注意してください：
- 文字は正確に読み取る
- 改行やレイアウトも可能な限り再現する
- 判読困難な部分は [判読不可] と記載する
- 日本語、英語、その他の言語すべて対応する

抽出されたテキスト:`

	case "translate":
		prompt = `この画像に含まれるテキストを抽出し、日本語に翻訳してください。
以下の点に注意してください：
- まず元のテキストを正確に抽出する
- その後、自然な日本語に翻訳する
- 元の言語が日本語の場合は、そのまま抽出結果を返す
- 専門用語は適切に翻訳する

抽出・翻訳結果:`

	case "summarize":
		prompt = `この画像に含まれるテキストを抽出し、内容を要約してください。
以下の点に注意してください：
- まず全文を読み取る
- 重要なポイントを3-5点で要約する
- 簡潔で分かりやすい日本語で記述する
- 元の文書の種類（メール、記事、レポート等）も判断する

要約結果:`

	case "analyze":
		prompt = `この画像に含まれるテキストや内容を詳細に分析してください。
以下の点を含めて分析してください：
- 文書の種類や目的
- 主要な内容のポイント
- 注意すべき点や重要な情報
- 文書の構造や形式
- その他気づいた点

詳細分析:`

	default:
		prompt = "この画像に含まれるテキストを抽出してください。"
	}

	// Luna AIとしてのプロンプト
	fullPrompt := fmt.Sprintf(`あなたは「Luna AI」です。Discord ボット「Luna」に統合されたOCR・画像解析機能として動作しています。

%s

ユーザーID: %s`, prompt, userID)

	// 画像とプロンプトで回答を生成
	resp, err := s.model.GenerateContent(ctx,
		genai.Text(fullPrompt),
		genai.ImageData(mimeType, imageData),
	)
	if err != nil {
		return "", fmt.Errorf("Gemini APIの呼び出しに失敗しました: %w", err)
	}

	// レスポンスの処理
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("Geminiからの応答がありません")
	}

	// 候補の存在を確認
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("AIからの回答候補がありません")
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
