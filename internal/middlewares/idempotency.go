package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Habeebamoo/intunel-backend/internal/store"
	"github.com/Habeebamoo/intunel-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func IdempotencyMiddleware(s *store.IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST requests
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Idempotency-Key header is required")
			c.Abort()
			return
		}

		// Check if key already exists
		cached, err := s.Get(c.Request.Context(), idempotencyKey)
		if err == nil && cached != "" {
			// Return cached response
			var cachedResp map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &cachedResp); err == nil {
				c.JSON(http.StatusOK, cachedResp)
				c.Abort()
				return
			}
		}

		// Capture response body
		blw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// Store response in Redis after processing
		if blw.status == http.StatusAccepted || blw.status == http.StatusOK {
			s.Set(c.Request.Context(), idempotencyKey, blw.body.String())
		}
	}
}

// bodyWriter captures the response body
type bodyWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *bodyWriter) ReadBody(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}