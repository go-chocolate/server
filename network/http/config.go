package http

import (
	"github.com/go-chocolate/server/basic"
)

type Config struct {
	basic.Config

	TLS       TLSConfig
	Cors      CorsConfig
	Logger    LoggerConfig
	RateLimit RateLimitConfig
}

type TLSConfig struct {
	Key  string
	Cert string
}

func (c TLSConfig) Valid() bool {
	return c.Key != "" && c.Cert != ""
}

type CorsConfig struct {
	Enable bool
}

type LoggerConfig struct {
	Enable bool
}

type RateLimitConfig struct {
	RateLimit int
}
