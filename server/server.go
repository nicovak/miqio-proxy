package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-proxy/config"
	"github.com/nicovak/miqio-proxy/internal/cache"
	"github.com/nicovak/miqio-proxy/internal/handler"
	"github.com/nicovak/miqio-proxy/internal/middleware"
	"github.com/nicovak/miqio-proxy/internal/r2"
	repo "github.com/nicovak/miqio-proxy/internal/repository"
	"go.uber.org/zap"
)

type Server struct {
	Logger   *zap.SugaredLogger
	Pool     *pgxpool.Pool
	Repo     *repository.PgxRepository
	R2Client *r2.Client
}

func Run(srv *Server) {
	conf := config.Get()
	defer srv.Pool.Close()

	tenantRepo := repo.NewTenantRepository(srv.Repo)
	redirectRepo := repo.NewRedirectRepository(srv.Repo)
	domainRepo := repo.NewDomainRepository(srv.Repo)

	store := &cache.Store{
		Tenants:   cache.NewTenantCache(tenantRepo),
		Redirects: cache.NewRedirectCache(redirectRepo),
		Domains:   cache.NewDomainCache(domainRepo),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := store.LoadAll(); err != nil {
		srv.Logger.Panicw("failed to load cache", zap.Error(err))
	}

	store.StartSync(ctx, conf.CacheSyncInterval)

	e := echo.New()

	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.AccessLog())

	handler.RegisterRoutes(e, srv.Pool, store, srv.R2Client, conf.MiqioPageDomain, conf.AppOrigin)

	sigCtx, sigCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	addr := ":" + conf.Port
	srv.Logger.Infow("server starting", zap.String("addr", addr))

	sc := echo.StartConfig{
		Address:         addr,
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: 10 * time.Second,
		OnShutdownError: func(err error) {
			srv.Logger.Errorw("shutdown error", zap.Error(err))
		},
	}

	if err := sc.Start(sigCtx, e); err != nil {
		srv.Logger.Panicw("server error", zap.Error(err))
	}

	srv.Logger.Infow("server stopped")
}
