package middleware

import "github.com/go-chocolate/server/network/http/internal/logger"

func Logger() Middleware {
	return logger.Logger(logger.WithIgnorePath("/__health__"))
}
