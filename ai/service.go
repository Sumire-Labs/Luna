package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/Sumire-Labs/Luna/config"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

type Service struct {
	config         *config.GoogleCloudConfig
	predictionClient *aiplatform.PredictionClient
	projectID      string
	location       string
}

func NewService(cfg *config.GoogleCloudConfig) (*Service, error) {
	ctx := context.Background()
	
	// 認証情報の設定
	var opts []option.ClientOption
	if cfg.CredentialsPath != "" {
		// 環境変数を設定
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.CredentialsPath)
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsPath))
	}
	
	// エンドポイントの設定
	endpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", cfg.Location)
	opts = append(opts, option.WithEndpoint(endpoint))
	
	// Prediction クライアントの作成
	predictionClient, err := aiplatform.NewPredictionClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prediction client: %w", err)
	}
	
	return &Service{
		config:           cfg,
		predictionClient: predictionClient,
		projectID:        cfg.ProjectID,
		location:         cfg.Location,
	}, nil
}

// AskGemini はGeminiモデルに質問を送信して回答を取得します
func (s *Service) AskGemini(ctx context.Context, question string, userID string) (string, error) {
	// タイムアウト設定
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// モデルのエンドポイント
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		s.projectID, s.location, s.config.GeminiModel)
	
	// Luna AI用の強化されたプロンプト
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
	
	// リクエストの構築
	promptValue, err := structpb.NewValue(map[string]interface{}{
		"contents": []interface{}{
			map[string]interface{}{
				"role": "user",
				"parts": []interface{}{
					map[string]interface{}{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.8,  // Gemini 2.5向けに調整
			"topP":           0.95,
			"topK":           64,    // より多様な回答のため増加
			"maxOutputTokens": 2048, // Gemini 2.5は出力が向上
		},
		"safetySettings": []interface{}{
			map[string]interface{}{
				"category":  "HARM_CATEGORY_HATE_SPEECH",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			map[string]interface{}{
				"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			map[string]interface{}{
				"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			map[string]interface{}{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	})
	
	if err != nil {
		return "", fmt.Errorf("プロンプトの構築に失敗しました: %w", err)
	}
	
	req := &aiplatformpb.PredictRequest{
		Endpoint:  endpoint,
		Instances: []*structpb.Value{promptValue},
	}
	
	// APIリクエスト
	resp, err := s.predictionClient.Predict(ctx, req)
	if err != nil {
		return "", fmt.Errorf("Gemini APIの呼び出しに失敗しました: %w", err)
	}
	
	// レスポンスの解析
	if len(resp.Predictions) == 0 {
		return "", fmt.Errorf("Geminiからの応答がありません")
	}
	
	// 回答を抽出
	prediction := resp.Predictions[0]
	if candidatesField, ok := prediction.GetStructValue().Fields["candidates"]; ok {
		if candidates := candidatesField.GetListValue(); candidates != nil && len(candidates.Values) > 0 {
			candidate := candidates.Values[0].GetStructValue()
			if contentField, ok := candidate.Fields["content"]; ok {
				if content := contentField.GetStructValue(); content != nil {
					if partsField, ok := content.Fields["parts"]; ok {
						if parts := partsField.GetListValue(); parts != nil && len(parts.Values) > 0 {
							part := parts.Values[0].GetStructValue()
							if textField, ok := part.Fields["text"]; ok {
								return strings.TrimSpace(textField.GetStringValue()), nil
							}
						}
					}
				}
			}
		}
	}
	
	return "", fmt.Errorf("レスポンスの解析に失敗しました")
}

// GenerateImage はImagenモデルで画像を生成します
func (s *Service) GenerateImage(ctx context.Context, prompt string, userID string) ([]byte, error) {
	// タイムアウト設定（画像生成は時間がかかる）
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	// モデルのエンドポイント
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		s.projectID, s.location, s.config.ImagenModel)
	
	// Imagen 4用の強化されたプロンプト（日本語対応向上）
	enhancedPrompt := fmt.Sprintf(`Generate a stunning, ultra-high-quality image based on this description: %s
Style: Digital art, highly detailed, vibrant colors, professional quality, 8K resolution
Imagen 4.0 enhanced: photorealistic details, perfect composition, cinematic lighting`, prompt)
	
	// リクエストの構築（Imagen 4の新機能対応）
	promptValue, err := structpb.NewValue(map[string]interface{}{
		"prompt": enhancedPrompt,
		"sampleCount": 1,
		"aspectRatio": "1:1",  // Imagen 4は16:9, 9:16なども対応
		"safetyFilterLevel": "block_some",
		"personGeneration": "allow_adult",
		"addWatermark": false,  // Imagen 4の新機能
		"language": "ja",       // 日本語サポート
	})
	
	if err != nil {
		return nil, fmt.Errorf("プロンプトの構築に失敗しました: %w", err)
	}
	
	req := &aiplatformpb.PredictRequest{
		Endpoint:  endpoint,
		Instances: []*structpb.Value{promptValue},
	}
	
	// APIリクエスト
	resp, err := s.predictionClient.Predict(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Imagen APIの呼び出しに失敗しました: %w", err)
	}
	
	// レスポンスの解析
	if len(resp.Predictions) == 0 {
		return nil, fmt.Errorf("Imagenからの応答がありません")
	}
	
	// 画像データを抽出
	prediction := resp.Predictions[0]
	if bytesBase64Field, ok := prediction.GetStructValue().Fields["bytesBase64Encoded"]; ok {
		imageData := bytesBase64Field.GetStringValue()
		// Base64デコード
		return decodeBase64(imageData)
	}
	
	return nil, fmt.Errorf("画像データの抽出に失敗しました")
}

// Close はリソースをクリーンアップします
func (s *Service) Close() error {
	if s.predictionClient != nil {
		return s.predictionClient.Close()
	}
	return nil
}

// Helper function to decode base64
func decodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}