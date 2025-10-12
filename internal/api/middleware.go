package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// CORSMiddleware - CORS handling
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthMiddleware - API Key based authentication
func AuthMiddleware() gin.HandlerFunc {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable must be set")
	}

	return func(c *gin.Context) {
		// Get API key from header
		requestKey := c.GetHeader("X-API-Key")

		// Also check Authorization header (Bearer token)
		if requestKey == "" {
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				requestKey = authHeader[7:]
			}
		}

		// For /trade endpoint, also check request body for apiKey (TradingView compatibility)
		if requestKey == "" && c.Request.Method == "POST" && c.FullPath() == "/api/trade" {
			// Read the body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore the body for the handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse JSON and extract apiKey
				var bodyMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
					if key, exists := bodyMap["apiKey"]; exists {
						if keyStr, ok := key.(string); ok {
							requestKey = keyStr
						}
					}
				}
			}
		}

		if requestKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Missing API key",
				"error":   "API key required in X-API-Key header, Authorization Bearer token, or apiKey field in request body for /trade endpoint",
			})
			c.Abort()
			return
		}

		if requestKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid API key",
				"error":   "The provided API key is invalid",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Rate Limiting Middleware
var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Allow 100 requests per minute per IP
			limiter = rate.NewLimiter(rate.Every(time.Minute/100), 100)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Cleanup old limiters periodically
func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			// Clear all limiters to free memory
			limiters = make(map[string]*rate.Limiter)
			mu.Unlock()
		}
	}()
}

// LoggerMiddleware - Request logging
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Log format
		c.Writer.Header().Set("X-Response-Time", duration.String())
		
		if statusCode >= 400 {
			c.Error(gin.Error{
				Err:  nil,
				Type: gin.ErrorTypePublic,
				Meta: gin.H{
					"method":   method,
					"path":     path,
					"status":   statusCode,
					"duration": duration.String(),
				},
			})
		}
	}
}

// RequestIDMiddleware - Request ID tracking
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Set("RequestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}
