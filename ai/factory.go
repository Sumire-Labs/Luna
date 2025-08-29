package ai

import (
	"fmt"
	"os"

	"github.com/Sumire-Labs/Luna/config"
)

// ServiceType represents the type of AI service
type ServiceType string

const (
	ServiceTypeGeminiStudio ServiceType = "gemini_studio"
	ServiceTypeVertexGemini ServiceType = "vertex_gemini"
	ServiceTypeVertexLegacy ServiceType = "vertex_legacy"
)

// Services holds all available AI services
type Services struct {
	GeminiStudio *GeminiStudioService
	VertexGemini *VertexGeminiService
	LegacyService *Service
	
	// Service availability flags
	HasGeminiStudio bool
	HasVertexGemini bool
	HasLegacyService bool
}

// Factory creates and manages AI services
type Factory struct {
	config *config.GoogleCloudConfig
}

// NewFactory creates a new AI service factory
func NewFactory(cfg *config.GoogleCloudConfig) *Factory {
	return &Factory{
		config: cfg,
	}
}

// CreateServices creates all configured AI services
func (f *Factory) CreateServices() (*Services, error) {
	services := &Services{}
	var lastError error

	// Determine which services to create based on configuration
	if f.shouldUseGeminiStudio() {
		if err := f.createGeminiStudio(services); err != nil {
			fmt.Printf("Warning: Failed to create Gemini Studio service: %v\n", err)
			lastError = err
		}
	}

	if f.shouldUseVertexAI() {
		// Create new Vertex Gemini API
		if err := f.createVertexGemini(services); err != nil {
			fmt.Printf("Warning: Failed to create Vertex Gemini service: %v\n", err)
			lastError = err
		}
		
		// Create legacy Vertex API for Imagen support
		if err := f.createLegacyService(services); err != nil {
			fmt.Printf("Warning: Failed to create legacy Vertex service: %v\n", err)
			// Don't set lastError here as this is optional
		}
	}

	// Check if at least one service was created
	if !services.HasGeminiStudio && !services.HasVertexGemini && !services.HasLegacyService {
		if lastError != nil {
			return nil, fmt.Errorf("no AI services could be initialized: %w", lastError)
		}
		return nil, fmt.Errorf("no AI services configured")
	}

	return services, nil
}

// shouldUseGeminiStudio determines if Gemini Studio should be used
func (f *Factory) shouldUseGeminiStudio() bool {
	// Use Gemini Studio if explicitly enabled OR if API key is provided
	return (f.config.UseStudioAPI && f.config.StudioAPIKey != "") ||
		   (f.config.StudioAPIKey != "" && f.config.ProjectID != "")
}

// shouldUseVertexAI determines if Vertex AI should be used
func (f *Factory) shouldUseVertexAI() bool {
	return f.config.ProjectID != ""
}

// createGeminiStudio creates the Gemini Studio service
func (f *Factory) createGeminiStudio(services *Services) error {
	if f.config.StudioAPIKey == "" {
		return fmt.Errorf("Gemini Studio API key not configured")
	}

	services.GeminiStudio = NewGeminiStudioService(
		f.config.StudioAPIKey,
		f.config.GeminiModel,
	)
	services.HasGeminiStudio = true
	fmt.Println("✓ Gemini Studio service initialized")
	return nil
}

// createVertexGemini creates the new Vertex AI Gemini service
func (f *Factory) createVertexGemini(services *Services) error {
	if f.config.ProjectID == "" {
		return fmt.Errorf("GCP project ID not configured")
	}

	// Set credentials if path is provided
	if f.config.CredentialsPath != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", f.config.CredentialsPath)
	}

	service, err := NewVertexGeminiService(f.config)
	if err != nil {
		return fmt.Errorf("failed to create Vertex Gemini service: %w", err)
	}

	services.VertexGemini = service
	services.HasVertexGemini = true
	fmt.Println("✓ Vertex AI Gemini service initialized")
	return nil
}

// createLegacyService creates the legacy Vertex AI service (for Imagen)
func (f *Factory) createLegacyService(services *Services) error {
	if f.config.ProjectID == "" {
		return fmt.Errorf("GCP project ID not configured")
	}

	service, err := NewService(f.config)
	if err != nil {
		return fmt.Errorf("failed to create legacy Vertex service: %w", err)
	}

	services.LegacyService = service
	services.HasLegacyService = true
	fmt.Println("✓ Legacy Vertex AI service initialized (Imagen support)")
	return nil
}

// GetPrimaryService returns the primary AI service for text generation
func (s *Services) GetPrimaryService() ServiceType {
	// Priority: VertexGemini > GeminiStudio > Legacy
	if s.HasVertexGemini {
		return ServiceTypeVertexGemini
	}
	if s.HasGeminiStudio {
		return ServiceTypeGeminiStudio
	}
	if s.HasLegacyService {
		return ServiceTypeVertexLegacy
	}
	return ""
}

// Close closes all active services
func (s *Services) Close() error {
	var firstError error

	if s.VertexGemini != nil {
		if err := s.VertexGemini.Close(); err != nil && firstError == nil {
			firstError = err
		}
	}

	if s.LegacyService != nil {
		if err := s.LegacyService.Close(); err != nil && firstError == nil {
			firstError = err
		}
	}

	// GeminiStudio doesn't have a Close method (uses HTTP client)

	return firstError
}