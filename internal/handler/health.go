package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-proxy/internal/r2"
)

type HealthHandler struct {
	pool     *pgxpool.Pool
	r2Client *r2.Client
}

func NewHealthHandler(pool *pgxpool.Pool, r2Client *r2.Client) *HealthHandler {
	return &HealthHandler{pool: pool, r2Client: r2Client}
}

func (h *HealthHandler) Liveness(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readiness(c *echo.Context) error {
	ctx := c.Request().Context()

	if err := h.pool.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"reason": "postgres unreachable",
		})
	}

	if err := h.r2Client.HeadBucket(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"reason": "r2 unreachable",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
