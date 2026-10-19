package middleware

import (
	"osush/internal/log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func LoggingMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		c.Set("logger", logger)

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		logger.Info("incoming request",
			log.String("request_id", requestID),
			log.String("method", method),
			log.String("path", path),
			log.String("client_ip", clientIP),
			log.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logFields := []log.Field{
			log.String("request_id", requestID),
			log.String("method", method),
			log.String("path", path),
			log.Int("status", statusCode),
			log.Duration("latency", latency),
			log.String("client_ip", clientIP),
			log.Int("response_size", c.Writer.Size()),
		}

		if claims, exists := c.Get("user"); exists {
			if userClaims, ok := claims.(map[string]interface{}); ok {
				if sub, ok := userClaims["sub"].(string); ok {
					logFields = append(logFields, log.String("user_id", sub))
				}
			}
		}

		if len(c.Errors) > 0 {
			logFields = append(logFields, log.String("errors", c.Errors.String()))
		}

		switch {
		case statusCode >= 500:
			logger.Error("request completed with server error", logFields...)
		case statusCode >= 400:
			logger.Warn("request completed with client error", logFields...)
		default:
			logger.Info("request completed successfully", logFields...)
		}
	}
}

func GetLogger(c *gin.Context) log.Logger {
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(log.Logger); ok {
			return l
		}
	}
	panic("logger not found in context")
}

func GetRequestID(c *gin.Context) string {
	if reqID, exists := c.Get("request_id"); exists {
		if id, ok := reqID.(string); ok {
			return id
		}
	}
	return ""
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
