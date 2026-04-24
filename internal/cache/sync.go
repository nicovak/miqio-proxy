package cache

import (
	"context"
	"time"

	"github.com/nicovak/miqio-go/logger"
	"go.uber.org/zap"
)

type Store struct {
	Tenants   *TenantCache
	Redirects *RedirectCache
	Domains   *DomainCache
}

func (s *Store) LoadAll() error {
	if err := s.Tenants.LoadFull(); err != nil {
		return err
	}
	if err := s.Redirects.LoadFull(); err != nil {
		return err
	}
	if err := s.Domains.LoadFull(); err != nil {
		return err
	}
	return nil
}

func (s *Store) StartSync(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				logger.Get().Infow("cache sync stopped")
				return
			case <-ticker.C:
				s.syncAll()
			}
		}
	}()

	logger.Get().Infow("cache sync started", zap.Duration("interval", interval))
}

func (s *Store) syncAll() {
	if err := s.Tenants.Sync(); err != nil {
		logger.Get().Errorw("tenant cache sync failed", zap.Error(err))
	}
	if err := s.Redirects.Sync(); err != nil {
		logger.Get().Errorw("redirect cache sync failed", zap.Error(err))
	}
	if err := s.Domains.Sync(); err != nil {
		logger.Get().Errorw("domain cache sync failed", zap.Error(err))
	}
}
