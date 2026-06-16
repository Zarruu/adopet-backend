package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:4173",
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL != "" {
		urls := strings.Split(frontendURL, ",")
		for _, u := range urls {
			u = strings.TrimSpace(u)
			if u != "" {
				allowedOrigins = append(allowedOrigins, u)
			}
		}
	}

	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow any localhost in development
			if strings.HasPrefix(origin, "http://localhost:") {
				return true
			}
			for _, o := range allowedOrigins {
				if o == origin {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

