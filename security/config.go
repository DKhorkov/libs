package security

import "time"

// JWTConfig is a config for creating and parsing Json Web Tokens.
type JWTConfig struct {
	SecretKey       string
	Algorithm       string
	RefreshTokenTTL time.Duration
	AccessTokenTTL  time.Duration
}

// Config is common security config.
type Config struct {
	HashCost int
	JWT      JWTConfig
}
