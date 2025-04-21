package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info().Str("method", c.Request.Method).
			Str("path", c.FullPath()).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Int("status", c.Writer.Status()).
			Msg("")
	}
}
