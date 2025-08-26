package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GeminiStudioService はGoogle AI Studio API用のサービス
type GeminiStudioService struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGeminiStudioService はGoogle AI Studio API用のサービスを作成
func NewGeminiStudioService(apiKey string, model string) *GeminiStudioService {
	if model == "" {
		model = "gemini-2.5-pro-latest" // 最新のGemini 2.5 Pro
	}
	
	return &GeminiStudioService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeminiRequest はAPI リクエストの構造体
type GeminiRequest struct {
	Contents []Content `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings []SafetySetting `json:"safetySettings,omitempty"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	TopK            int     `json:"topK"`
	TopP            float64 `json:"topP"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiResponse はAPIレスポンスの構造体
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
	Error      *APIError   `json:"error,omitempty"`
}

type Candidate struct {
	Content Content `json:"content"`
	FinishReason string `json:"finishReason"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// AskGemini はGoogle AI Studio APIを使用してGeminiに質問
func (s *GeminiStudioService) AskGemini(ctx context.Context, question string, userID string) (string, error) {
	// APIエンドポイント
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.model, s.apiKey)
	
	// リクエストボディの構築
	prompt := fmt.Sprintf(`あなたは「Luna AI」です。Discord ボット「Luna」に統合された高性能AIアシスタントとして動作しています。

以下のガイドラインに従って回答してください：
- 親切で知識豊富なLuna AIとして振る舞う
- 日本語で丁寧に回答する
- Discord用に最適化された回答（2000文字以内）
- 最新の情報（2025年）に基づいて回答
- 必要に応じて絵文字を使って親しみやすく
- 自分を「Luna AI」または「私」と呼ぶ
- Geminiという名前は一切使わない

ユーザーID: %s
ユーザーの質問: %s`, userID, question)
	
	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
				Role: "user",
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature:     0.8,
			TopK:            64,
			TopP:            0.95,
			MaxOutputTokens: 2048,
		},
		SafetySettings: []SafetySetting{
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	}
	
	// JSONにエンコード
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("リクエストのエンコードに失敗: %w", err)
	}
	
	// HTTPリクエストの作成
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("リクエストの作成に失敗: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// APIリクエストの送信
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("APIリクエストに失敗: %w", err)
	}
	defer resp.Body.Close()
	
	// レスポンスの読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンスの読み取りに失敗: %w", err)
	}
	
	// JSONのデコード
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("レスポンスのデコードに失敗: %w", err)
	}
	
	// エラーチェック
	if geminiResp.Error != nil {
		return "", fmt.Errorf("API エラー: %s", geminiResp.Error.Message)
	}
	
	// レスポンスの確認
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("有効な回答が得られませんでした")
	}
	
	// テキストの抽出
	answer := geminiResp.Candidates[0].Content.Parts[0].Text
	
	return answer, nil
}

// GetAvailableModels は利用可能なモデルのリスト
func GetAvailableModels() []string {
	return []string{
		"gemini-2.5-pro-latest",     // Gemini 2.5 Pro 最新版
		"gemini-2.5-flash-latest",   // Gemini 2.5 Flash 最新版
		"gemini-1.5-pro-latest",     // Gemini 1.5 Pro
		"gemini-1.5-flash-latest",   // Gemini 1.5 Flash
		"gemini-1.5-flash-8b-latest", // Gemini 1.5 Flash 8B
	}
}

// GetFreeQuota は無料枠の情報
func GetFreeQuota() string {
	return `
Google AI Studio API 無料枠（2025年1月現在）:
- リクエスト数: 15 RPM (Requests Per Minute)
- 1日あたり: 1,500リクエスト
- トークン制限: 1分あたり100万トークン
- 料金: 完全無料！

注意: 商用利用の場合はVertex AIの使用を推奨
`
}