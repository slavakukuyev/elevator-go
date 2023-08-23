package main

type Config struct {
	Port     int `env:"PORT" envDefault:"1010"`
	MaxFloor int `env:"DEFAULT_MAX_FLOOR" envDefault:"9"`
	MinFloor int `env:"DEFAULT_MIN_FLOOR" envDefault:"0"`
}

var cfg Config

const _directionUp = "up"
const _directionDown = "down"
