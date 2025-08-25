package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MultimodalRequest はマルチモーダル対応のリクエスト構造体
type MultimodalRequest struct {
	Contents         []MultimodalContent `json:"contents"`
	GenerationConfig GenerationConfig    `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting     `json:"safetySettings,omitempty"`
}

type MultimodalContent struct {
	Parts []MultimodalPart `json:"parts"`
	Role  string           `json:"role"`
}

type MultimodalPart struct {
	Text       string     `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64エンコードされた画像データ
}

// OCRWithGemini は画像からテキストを抽出します
func (s *GeminiStudioService) OCRWithGemini(ctx context.Context, imageData []byte, mimeType string, userID string, extractType string) (string, error) {
	// 画像データをBase64エンコード
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	
	// プロンプトの構築（タスクに応じて変更）
	var prompt string
	switch extractType {
	case "text":
		prompt = `この画像に含まれているテキストを正確に読み取って、そのまま出力してください。
文字の配置や改行も可能な限り再現してください。
テキストが見つからない場合は「テキストが見つかりませんでした」と回答してください。`
	
	case "translate":
		prompt = `この画像に含まれているテキストを読み取り、日本語に翻訳してください。
元のテキストと翻訳結果の両方を出力してください。
形式：
【元のテキスト】
（抽出されたテキスト）

【日本語翻訳】
（翻訳結果）`
	
	case "summarize":
		prompt = `この画像に含まれているテキストを読み取り、内容を要約してください。
以下の形式で出力してください：
【抽出テキスト】
（元のテキスト）

【要約】
（要約内容）

【キーポイント】
・主要な点1
・主要な点2
・主要な点3`
	
	case "analyze":
		prompt = `この画像を詳細に分析して以下の情報を教えてください：
1. 含まれているテキスト（OCR）
2. 画像の内容・構成
3. テキストの種類（文書、看板、メニューなど）
4. 特徴的な要素

日本語で分かりやすく回答してください。`
	
	default:
		prompt = `この画像に含まれているテキストを正確に読み取って出力してください。
読みやすい形式で整理して表示してください。`
	}
	
	// APIエンドポイント
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.model, s.apiKey)
	
	// リクエストボディの構築
	request := MultimodalRequest{
		Contents: []MultimodalContent{
			{
				Parts: []MultimodalPart{
					{
						Text: prompt,
					},
					{
						InlineData: &InlineData{
							MimeType: mimeType,
							Data:     base64Image,
						},
					},
				},
				Role: "user",
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature:     0.4, // OCRには低い温度が適している
			TopK:            32,
			TopP:            0.8,
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
	result := geminiResp.Candidates[0].Content.Parts[0].Text
	
	return strings.TrimSpace(result), nil
}

// DownloadImage はURLから画像をダウンロードします
func DownloadImage(ctx context.Context, url string) ([]byte, string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("リクエストの作成に失敗: %w", err)
	}
	
	req.Header.Set("User-Agent", "Luna-Discord-Bot/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("画像のダウンロードに失敗: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("画像のダウンロードに失敗: %s", resp.Status)
	}
	
	// Content-Typeを取得
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	
	// サポートされている画像形式をチェック
	supportedTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif", 
		"image/webp", "image/bmp", "image/svg+xml",
	}
	
	isSupported := false
	for _, supportedType := range supportedTypes {
		if strings.Contains(mimeType, supportedType) || mimeType == supportedType {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		return nil, "", fmt.Errorf("サポートされていない画像形式: %s", mimeType)
	}
	
	// 画像データを読み取り
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("画像データの読み取りに失敗: %w", err)
	}
	
	// ファイルサイズチェック（20MB制限）
	const maxSize = 20 * 1024 * 1024 // 20MB
	if len(data) > maxSize {
		return nil, "", fmt.Errorf("画像ファイルが大きすぎます (最大: 20MB)")
	}
	
	return data, mimeType, nil
}

// GetSupportedImageTypes はサポートされている画像形式のリストを返します
func GetSupportedImageTypes() []string {
	return []string{
		"JPEG", "JPG", "PNG", "GIF", "WebP", "BMP", "SVG",
	}
}