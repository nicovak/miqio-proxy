package middleware

import (
	"time"

	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-go/logger"
	"go.uber.org/zap"
)

func AccessLog() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res, _ := echo.UnwrapResponse(c.Response())

			logger.Get().Infow("request",
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("host", req.Host),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.Duration("latency", time.Since(start)),
				zap.String("remote_ip", c.RealIP()),
				zap.String("request_id", res.Header().Get(echo.HeaderXRequestID)),
			)

			return err
		}
	}
}
