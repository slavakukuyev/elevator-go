package config

import (
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	Port int `env:"PORT" envDefault:"6660"`

	MaxFloor int `env:"DEFAULT_MAX_FLOOR" envDefault:"9"`
	MinFloor int `env:"DEFAULT_MIN_FLOOR" envDefault:"0"`

	EachFloorDuration time.Duration `env:"EACH_FLOOR_DURATION" envDefault:"500ms"`
	OpenDoorDuration  time.Duration `env:"OPEN_DOOR_DURATION" envDefault:"2s"`

	DirectionUpKey   string
	DirectionDownKey string
}

func InitConfig() *Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		panic("error on parsing env")
	}

	cfg.DirectionUpKey = "up"
	cfg.DirectionDownKey = "down"

	return &cfg
}
