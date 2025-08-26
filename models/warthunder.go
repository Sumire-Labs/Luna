package models

import "time"

// GameMode represents the War Thunder game modes
type GameMode int

const (
	GameModeAir GameMode = iota
	GameModeGround
	GameModeNaval
)

func (gm GameMode) String() string {
	switch gm {
	case GameModeAir:
		return "Air"
	case GameModeGround:
		return "Ground"
	case GameModeNaval:
		return "Naval"
	default:
		return "Unknown"
	}
}

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

// Nation represents the War Thunder nations
type Nation string

const (
	NationUSA     Nation = "USA"
	NationGermany Nation = "Germany"
	NationUSSR    Nation = "USSR"
	NationBritain Nation = "Britain"
	NationJapan   Nation = "Japan"
	NationChina   Nation = "China"
	NationItaly   Nation = "Italy"
	NationFrance  Nation = "France"
	NationSweden  Nation = "Sweden"
	NationIsrael  Nation = "Israel"
)

func (n Nation) Flag() string {
	switch n {
	case NationUSA:
		return "🇺🇸"
	case NationGermany:
		return "🇩🇪"
	case NationUSSR:
		return "🇷🇺"
	case NationBritain:
		return "🇬🇧"
	case NationJapan:
		return "🇯🇵"
	case NationChina:
		return "🇨🇳"
	case NationItaly:
		return "🇮🇹"
	case NationFrance:
		return "🇫🇷"
	case NationSweden:
		return "🇸🇪"
	case NationIsrael:
		return "🇮🇱"
	default:
		return "🏳️"
	}
}

// VehicleType represents the type of vehicle
type VehicleType string

const (
	// Air
	VehicleTypeFighter      VehicleType = "Fighter"
	VehicleTypeAttacker     VehicleType = "Attacker"
	VehicleTypeBomber       VehicleType = "Bomber"
	VehicleTypeHelicopter   VehicleType = "Helicopter"
	
	// Ground
	VehicleTypeLightTank    VehicleType = "Light Tank"
	VehicleTypeMediumTank   VehicleType = "Medium Tank"
	VehicleTypeHeavyTank    VehicleType = "Heavy Tank"
	VehicleTypeTankDestroyer VehicleType = "Tank Destroyer"
	VehicleTypeSPAA         VehicleType = "SPAA"
	VehicleTypeSPG          VehicleType = "SPG"
	
	// Naval
	VehicleTypeBoat         VehicleType = "Boat"
	VehicleTypeFrigate      VehicleType = "Frigate"
	VehicleTypeDestroyer    VehicleType = "Destroyer"
	VehicleTypeCruiser      VehicleType = "Cruiser"
	VehicleTypeBattleship   VehicleType = "Battleship"
	VehicleTypeCarrier      VehicleType = "Carrier"
	VehicleTypeSubmarine    VehicleType = "Submarine"
)

// Vehicle represents a War Thunder vehicle
type Vehicle struct {
	ID          int         `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Nation      Nation      `json:"nation" db:"nation"`
	GameMode    GameMode    `json:"game_mode" db:"game_mode"`
	VehicleType VehicleType `json:"vehicle_type" db:"vehicle_type"`
	BR          float64     `json:"br" db:"br"`
	Rank        int         `json:"rank" db:"rank"`
	Premium     bool        `json:"premium" db:"premium"`
	Squadron    bool        `json:"squadron" db:"squadron"`
	Event       bool        `json:"event" db:"event"`
	ImageURL    string      `json:"image_url" db:"image_url"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// RouletteConfig represents user's roulette configuration
type RouletteConfig struct {
	ID              int      `json:"id" db:"id"`
	UserID          string   `json:"user_id" db:"user_id"`
	GuildID         string   `json:"guild_id" db:"guild_id"`
	GameMode        GameMode `json:"game_mode" db:"game_mode"`
	MinBR           float64  `json:"min_br" db:"min_br"`
	MaxBR           float64  `json:"max_br" db:"max_br"`
	ExcludedNations string   `json:"excluded_nations" db:"excluded_nations"` // JSON array
	ExcludedTypes   string   `json:"excluded_types" db:"excluded_types"`     // JSON array
	ExcludedVehicles string  `json:"excluded_vehicles" db:"excluded_vehicles"` // JSON array
	IncludePremium  bool     `json:"include_premium" db:"include_premium"`
	IncludeSquadron bool     `json:"include_squadron" db:"include_squadron"`
	IncludeEvent    bool     `json:"include_event" db:"include_event"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RouletteResult represents the result of a roulette spin
type RouletteResult struct {
	Vehicle   *Vehicle       `json:"vehicle"`
	Config    *RouletteConfig `json:"config"`
	Timestamp time.Time      `json:"timestamp"`
}