package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Sumire-Labs/Luna/models"
)

type WarThunderService struct {
	db *sql.DB
}

func NewWarThunderService(db *sql.DB) *WarThunderService {
	return &WarThunderService{db: db}
}

// InitializeTables creates the necessary tables for War Thunder functionality
func (wts *WarThunderService) InitializeTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS wt_vehicles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			nation TEXT NOT NULL,
			game_mode INTEGER NOT NULL,
			vehicle_type TEXT NOT NULL,
			br REAL NOT NULL,
			rank INTEGER NOT NULL,
			premium BOOLEAN DEFAULT FALSE,
			squadron BOOLEAN DEFAULT FALSE,
			event BOOLEAN DEFAULT FALSE,
			image_url TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_wt_vehicles_game_mode ON wt_vehicles(game_mode)`,
		`CREATE INDEX IF NOT EXISTS idx_wt_vehicles_br ON wt_vehicles(br)`,
		`CREATE INDEX IF NOT EXISTS idx_wt_vehicles_nation ON wt_vehicles(nation)`,
		`CREATE TABLE IF NOT EXISTS wt_roulette_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			guild_id TEXT NOT NULL,
			game_mode INTEGER NOT NULL,
			min_br REAL DEFAULT 1.0,
			max_br REAL DEFAULT 13.0,
			excluded_nations TEXT DEFAULT '[]',
			excluded_types TEXT DEFAULT '[]',
			excluded_vehicles TEXT DEFAULT '[]',
			include_premium BOOLEAN DEFAULT TRUE,
			include_squadron BOOLEAN DEFAULT TRUE,
			include_event BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, guild_id, game_mode)
		)`,
	}

	for _, query := range queries {
		if _, err := wts.db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// GetVehicles retrieves vehicles based on filters
func (wts *WarThunderService) GetVehicles(gameMode models.GameMode, minBR, maxBR float64, config *models.RouletteConfig) ([]*models.Vehicle, error) {
	query := `
		SELECT id, name, nation, game_mode, vehicle_type, br, rank, premium, squadron, event, image_url, created_at, updated_at
		FROM wt_vehicles 
		WHERE game_mode = ? AND br >= ? AND br <= ?
	`
	args := []interface{}{gameMode, minBR, maxBR}

	// Apply exclusion filters if config is provided
	if config != nil {
		// Exclude premium vehicles if needed
		if !config.IncludePremium {
			query += " AND premium = FALSE"
		}
		
		// Exclude squadron vehicles if needed
		if !config.IncludeSquadron {
			query += " AND squadron = FALSE"
		}
		
		// Exclude event vehicles if needed
		if !config.IncludeEvent {
			query += " AND event = FALSE"
		}

		// Handle excluded nations
		var excludedNations []string
		if config.ExcludedNations != "" && config.ExcludedNations != "[]" {
			json.Unmarshal([]byte(config.ExcludedNations), &excludedNations)
			if len(excludedNations) > 0 {
				placeholders := strings.Repeat("?,", len(excludedNations)-1) + "?"
				query += fmt.Sprintf(" AND nation NOT IN (%s)", placeholders)
				for _, nation := range excludedNations {
					args = append(args, nation)
				}
			}
		}

		// Handle excluded types
		var excludedTypes []string
		if config.ExcludedTypes != "" && config.ExcludedTypes != "[]" {
			json.Unmarshal([]byte(config.ExcludedTypes), &excludedTypes)
			if len(excludedTypes) > 0 {
				placeholders := strings.Repeat("?,", len(excludedTypes)-1) + "?"
				query += fmt.Sprintf(" AND vehicle_type NOT IN (%s)", placeholders)
				for _, vType := range excludedTypes {
					args = append(args, vType)
				}
			}
		}

		// Handle excluded vehicles
		var excludedVehicles []int
		if config.ExcludedVehicles != "" && config.ExcludedVehicles != "[]" {
			json.Unmarshal([]byte(config.ExcludedVehicles), &excludedVehicles)
			if len(excludedVehicles) > 0 {
				placeholders := strings.Repeat("?,", len(excludedVehicles)-1) + "?"
				query += fmt.Sprintf(" AND id NOT IN (%s)", placeholders)
				for _, id := range excludedVehicles {
					args = append(args, id)
				}
			}
		}
	}

	query += " ORDER BY br, name"

	rows, err := wts.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []*models.Vehicle
	for rows.Next() {
		vehicle := &models.Vehicle{}
		err := rows.Scan(
			&vehicle.ID, &vehicle.Name, &vehicle.Nation, &vehicle.GameMode,
			&vehicle.VehicleType, &vehicle.BR, &vehicle.Rank, &vehicle.Premium,
			&vehicle.Squadron, &vehicle.Event, &vehicle.ImageURL,
			&vehicle.CreatedAt, &vehicle.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle: %w", err)
		}
		vehicles = append(vehicles, vehicle)
	}

	return vehicles, nil
}

// SpinRoulette performs a roulette spin and returns a random vehicle
func (wts *WarThunderService) SpinRoulette(config *models.RouletteConfig) (*models.RouletteResult, error) {
	vehicles, err := wts.GetVehicles(config.GameMode, config.MinBR, config.MaxBR, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get vehicles for roulette: %w", err)
	}

	if len(vehicles) == 0 {
		return nil, fmt.Errorf("no vehicles found matching the criteria")
	}

	// Select random vehicle
	rand.Seed(time.Now().UnixNano())
	selectedVehicle := vehicles[rand.Intn(len(vehicles))]

	result := &models.RouletteResult{
		Vehicle:   selectedVehicle,
		Config:    config,
		Timestamp: time.Now(),
	}

	return result, nil
}

// GetRouletteConfig retrieves user's roulette configuration
func (wts *WarThunderService) GetRouletteConfig(userID, guildID string, gameMode models.GameMode) (*models.RouletteConfig, error) {
	query := `
		SELECT id, user_id, guild_id, game_mode, min_br, max_br, excluded_nations, excluded_types, excluded_vehicles,
		       include_premium, include_squadron, include_event, created_at, updated_at
		FROM wt_roulette_configs
		WHERE user_id = ? AND guild_id = ? AND game_mode = ?
	`

	config := &models.RouletteConfig{}
	err := wts.db.QueryRow(query, userID, guildID, gameMode).Scan(
		&config.ID, &config.UserID, &config.GuildID, &config.GameMode,
		&config.MinBR, &config.MaxBR, &config.ExcludedNations, &config.ExcludedTypes,
		&config.ExcludedVehicles, &config.IncludePremium, &config.IncludeSquadron,
		&config.IncludeEvent, &config.CreatedAt, &config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default config if none exists
		return wts.createDefaultConfig(userID, guildID, gameMode)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get roulette config: %w", err)
	}

	return config, nil
}

// SaveRouletteConfig saves user's roulette configuration
func (wts *WarThunderService) SaveRouletteConfig(config *models.RouletteConfig) error {
	query := `
		INSERT OR REPLACE INTO wt_roulette_configs 
		(user_id, guild_id, game_mode, min_br, max_br, excluded_nations, excluded_types, excluded_vehicles,
		 include_premium, include_squadron, include_event, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := wts.db.Exec(query,
		config.UserID, config.GuildID, config.GameMode, config.MinBR, config.MaxBR,
		config.ExcludedNations, config.ExcludedTypes, config.ExcludedVehicles,
		config.IncludePremium, config.IncludeSquadron, config.IncludeEvent,
	)

	if err != nil {
		return fmt.Errorf("failed to save roulette config: %w", err)
	}

	return nil
}

// createDefaultConfig creates a default configuration for the user
func (wts *WarThunderService) createDefaultConfig(userID, guildID string, gameMode models.GameMode) (*models.RouletteConfig, error) {
	config := &models.RouletteConfig{
		UserID:           userID,
		GuildID:          guildID,
		GameMode:         gameMode,
		MinBR:           1.0,
		MaxBR:           13.0,
		ExcludedNations:  "[]",
		ExcludedTypes:    "[]",
		ExcludedVehicles: "[]",
		IncludePremium:   true,
		IncludeSquadron:  true,
		IncludeEvent:     false,
	}

	// Adjust default BR ranges based on game mode
	switch gameMode {
	case models.GameModeAir:
		config.MaxBR = 13.0
	case models.GameModeGround:
		config.MaxBR = 12.0
	case models.GameModeNaval:
		config.MaxBR = 7.0
	}

	err := wts.SaveRouletteConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// AddVehicle adds a new vehicle to the database
func (wts *WarThunderService) AddVehicle(vehicle *models.Vehicle) error {
	query := `
		INSERT INTO wt_vehicles (name, nation, game_mode, vehicle_type, br, rank, premium, squadron, event, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := wts.db.Exec(query,
		vehicle.Name, vehicle.Nation, vehicle.GameMode, vehicle.VehicleType,
		vehicle.BR, vehicle.Rank, vehicle.Premium, vehicle.Squadron,
		vehicle.Event, vehicle.ImageURL,
	)

	if err != nil {
		return fmt.Errorf("failed to add vehicle: %w", err)
	}

	return nil
}

// GetVehicleCount returns the total number of vehicles in the database
func (wts *WarThunderService) GetVehicleCount(gameMode *models.GameMode) (int, error) {
	var query string
	var args []interface{}

	if gameMode != nil {
		query = "SELECT COUNT(*) FROM wt_vehicles WHERE game_mode = ?"
		args = append(args, *gameMode)
	} else {
		query = "SELECT COUNT(*) FROM wt_vehicles"
	}

	var count int
	err := wts.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get vehicle count: %w", err)
	}

	return count, nil
}