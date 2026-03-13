package handler

import "net/http"

// HealthCheck -- эндпоинт проверки здоровья сервиса.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}
