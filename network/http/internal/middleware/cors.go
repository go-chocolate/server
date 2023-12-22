package middleware

import (
	"github.com/go-chocolate/server/network/http/internal/cors"
)

type CORSOption func(c *cors.Config)

func applyCORSOptions(options ...CORSOption) cors.Config {
	c := cors.DefaultConfig()
	for _, option := range options {
		option(&c)
	}
	return c
}

func CORS(options ...CORSOption) Middleware {
	return cors.New(applyCORSOptions(options...))
}
