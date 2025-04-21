package logger

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log.Logger = log.Output(&buf)

	r := gin.New()
	r.Use(Middleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	out := buf.String()
	assert.Contains(t, out, `"method":"GET"`)
	assert.Contains(t, out, `"path":"/test"`)
	assert.Contains(t, out, `"status":200`)
	assert.Contains(t, out, `"duration_ms"`)
}
