package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redisClient *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
	}
}

// EmailVerificationRateLimit limits email verification requests
// Allows 3 requests per email per 10 minutes
func (rl *RateLimiter) EmailVerificationRateLimit() gin.HandlerFunc {
	return rl.RateLimit("email_verification", 3, 10*time.Minute, "email")
}

// PasswordResetRateLimit limits password reset requests
// Allows 2 requests per email per 15 minutes
func (rl *RateLimiter) PasswordResetRateLimit() gin.HandlerFunc {
	return rl.RateLimit("password_reset", 2, 15*time.Minute, "email")
}

// RateLimit creates a generic rate limiting middleware
func (rl *RateLimiter) RateLimit(keyPrefix string, maxRequests int, window time.Duration, identifierKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier from request body or query params
		identifier := rl.getIdentifier(c, identifierKey)
		if identifier == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing identifier for rate limiting"})
			c.Abort()
			return
		}

		// Create rate limiting key
		key := fmt.Sprintf("rate_limit:%s:%s", keyPrefix, identifier)

		ctx := context.Background()

		// Get current count
		current, err := rl.redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// If Redis is down, allow the request but log the error
			// In production, you might want to handle this differently
			c.Next()
			return
		}

		if current >= maxRequests {
			// Get TTL to inform user when they can try again
			ttl, _ := rl.redisClient.TTL(ctx, key).Result()

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"message":     fmt.Sprintf("Rate limit exceeded. Try again in %v", ttl.Round(time.Second)),
				"retry_after": int(ttl.Seconds()),
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := rl.redisClient.Pipeline()
		pipe.Incr(ctx, key)
		if current == 0 {
			// Set expiration only on first request
			pipe.Expire(ctx, key, window)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			// If Redis is down, allow the request but log the error
			c.Next()
			return
		}

		// Add headers to inform client about rate limiting
		remaining := maxRequests - current - 1
		if remaining < 0 {
			remaining = 0
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Window", window.String())

		c.Next()
	}
}

// getIdentifier extracts the identifier from the request
func (rl *RateLimiter) getIdentifier(c *gin.Context, identifierKey string) string {
	// First try to get from request body (for POST requests)
	if c.Request.Method == "POST" {
		// Read the request body without consuming it
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 {
				// Restore the request body for the next handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Try to parse as JSON
				var requestBody map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &requestBody); err == nil {
					if value, exists := requestBody[identifierKey]; exists {
						if str, ok := value.(string); ok {
							return strings.ToLower(str)
						}
					}
				}
			}
		}
	}

	// Fallback to query parameter
	value := c.Query(identifierKey)
	if value != "" {
		return strings.ToLower(value)
	}

	// Fallback to form parameter
	value = c.PostForm(identifierKey)
	if value != "" {
		return strings.ToLower(value)
	}

	return ""
}
