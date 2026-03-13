package middleware

import (
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RequestLogger логирует входящие HTTP-запросы.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			zap.L().Info("HTTP запрос",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Int64("duration_ms", time.Since(start).Milliseconds()),
				zap.String("request_id", chimw.GetReqID(r.Context())),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
