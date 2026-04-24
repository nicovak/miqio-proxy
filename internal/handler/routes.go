package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-proxy/internal/cache"
	"github.com/nicovak/miqio-proxy/internal/r2"
)

func RegisterRoutes(e *echo.Echo, pool *pgxpool.Pool, store *cache.Store, r2Client *r2.Client, miqioPageDomain, appOrigin string) {
	health := NewHealthHandler(pool, r2Client)
	serve := NewServeHandler(store, r2Client, miqioPageDomain)
	track := NewTrackHandler(store, miqioPageDomain, appOrigin)

	e.GET("/healthz", health.Liveness)
	e.GET("/readyz", health.Readiness)

	e.POST("/api/track", track.Handle)

	e.GET("/", serve.Handle)
	e.GET("/*", serve.Handle)
}
