package services

import (
	"fmt"
	"math/rand"
	"time"
)

// GameMode represents War Thunder game modes
type GameMode string

const (
	GameModeAir    GameMode = "air"
	GameModeGround GameMode = "ground"
	GameModeNaval  GameMode = "naval"
)

func (gm GameMode) Emoji() string {
	switch gm {
	case GameModeAir:
		return "🛩️"
	case GameModeGround:
		return "🚗"
	case GameModeNaval:
		return "🚢"
	default:
		return "❓"
	}
}

// WarThunderSimpleService provides simple BR roulette functionality
type WarThunderSimpleService struct {
	groundBRs []float64
	airBRs    []float64
	navalBRs  []float64
}

// NewWarThunderSimpleService creates a new War Thunder service
func NewWarThunderSimpleService() *WarThunderSimpleService {
	return &WarThunderSimpleService{
		groundBRs: generateBRList(1.0, 12.0),
		airBRs:    generateBRList(1.0, 14.0),
		navalBRs:  generateBRList(1.0, 8.7),
	}
}

// generateBRList generates the correct BR progression pattern
// Pattern: 1.0, 1.3, 1.7, 2.0, 2.3, 2.7, 3.0, 3.3, 3.7, 4.0...
func generateBRList(minBR, maxBR float64) []float64 {
	brs := []float64{}
	
	// Start from 1.0
	current := 1.0
	
	for current <= maxBR {
		brs = append(brs, current)
		
		// Calculate next BR
		// Pattern: x.0 -> x.3 -> x.7 -> (x+1).0
		fraction := current - float64(int(current))
		
		if fraction < 0.1 { // x.0
			current += 0.3
		} else if fraction < 0.4 { // x.3
			current += 0.4
		} else { // x.7
			current = float64(int(current) + 1) // Next whole number
		}
		
		// Round to 1 decimal place to avoid floating point errors
		current = float64(int(current*10+0.5)) / 10
	}
	
	return brs
}

// GetRandomBR returns a random BR for the specified game mode
func (wts *WarThunderSimpleService) GetRandomBR(gameMode GameMode, minBR, maxBR float64) (float64, error) {
	var availableBRs []float64
	
	// Select BR list based on game mode
	switch gameMode {
	case GameModeGround:
		availableBRs = wts.groundBRs
		if maxBR > 12.0 {
			maxBR = 12.0
		}
	case GameModeAir:
		availableBRs = wts.airBRs
		if maxBR > 14.0 {
			maxBR = 14.0
		}
	case GameModeNaval:
		availableBRs = wts.navalBRs
		if maxBR > 8.7 {
			maxBR = 8.7
		}
	default:
		return 0, fmt.Errorf("invalid game mode: %s", gameMode)
	}
	
	// Filter BRs within specified range
	filteredBRs := []float64{}
	for _, br := range availableBRs {
		if br >= minBR && br <= maxBR {
			filteredBRs = append(filteredBRs, br)
		}
	}
	
	if len(filteredBRs) == 0 {
		return 0, fmt.Errorf("no valid BRs in range %.1f - %.1f for %s", minBR, maxBR, gameMode)
	}
	
	// Select random BR
	rand.Seed(time.Now().UnixNano())
	return filteredBRs[rand.Intn(len(filteredBRs))], nil
}

// GetDefaultBRRange returns the default BR range for each game mode
func (wts *WarThunderSimpleService) GetDefaultBRRange(gameMode GameMode) (float64, float64) {
	switch gameMode {
	case GameModeGround:
		return 1.0, 12.0
	case GameModeAir:
		return 1.0, 14.0
	case GameModeNaval:
		return 1.0, 8.7
	default:
		return 1.0, 12.0
	}
}