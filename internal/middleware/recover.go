package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-go/logger"
	"go.uber.org/zap"
)

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)

				logger.Get().Errorw("panic recovered",
					zap.String("error", fmt.Sprintf("%v", r)),
					zap.String("stack", string(buf[:n])),
					zap.String("uri", c.Request().RequestURI),
				)

				_ = c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
				}
			}()
			return next(c)
		}
	}
}
