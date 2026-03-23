package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"

	"booking-service/app/api/dto"
)

// Recovery восстанавливает после panic и возвращает ProblemDetails.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				zap.L().Error("panic восстановлен",
					zap.Any("panic", rec),
					zap.String("stack", string(debug.Stack())),
				)

				pd := dto.ProblemDetails{
					Type:   "about:blank",
					Title:  "Внутренняя ошибка сервера",
					Status: http.StatusInternalServerError,
				}

				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(pd)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
