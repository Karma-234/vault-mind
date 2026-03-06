package middleware

import (
	"log"
	"net/http"
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// // rateLimit := rate.NewLimiter(1, 5) // 1 request per second with a burst of 5
		// if !rateLimit.Allow() {
		// 	http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		// 	return
		// }
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
