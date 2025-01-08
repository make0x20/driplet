package middleware

import (
	"net/http"
	"log/slog"
)

// Chain applies middlewares in order
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

// DefaultChain middleware chain
func DefaultChain(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(h http.Handler) http.Handler {
        return Chain(h,
            LoggerMiddleware(logger),
        )
    }
}
