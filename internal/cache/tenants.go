package cache

import (
	"sync"
	"time"

	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-proxy/internal/repository"
	"go.uber.org/zap"
)

type TenantCache struct {
	mu   sync.RWMutex
	data map[string]string // tenantSlug -> tenantUUID
	repo repository.TenantRepo
	last time.Time
}

func NewTenantCache(repo repository.TenantRepo) *TenantCache {
	return &TenantCache{
		data: make(map[string]string),
		repo: repo,
	}
}

func (c *TenantCache) LoadFull() error {
	rows, err := c.repo.GetActiveTenants()
	if err != nil {
		return err
	}

	data := make(map[string]string, len(rows))
	for _, r := range rows {
		data[r.Slug] = r.ID
	}

	c.mu.Lock()
	c.data = data
	c.last = time.Now()
	c.mu.Unlock()

	logger.Get().Infow("tenant cache loaded", zap.Int("count", len(data)))
	return nil
}

func (c *TenantCache) Sync() error {
	c.mu.RLock()
	since := c.last
	c.mu.RUnlock()

	rows, err := c.repo.GetUpdatedTenants(since)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range rows {
		if r.Status == "active" {
			c.data[r.Slug] = r.ID
		} else {
			delete(c.data, r.Slug)
		}
	}

	c.last = time.Now()

	if len(rows) > 0 {
		logger.Get().Infow("tenant cache synced", zap.Int("updated", len(rows)))
	}
	return nil
}

func (c *TenantCache) Resolve(slug string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	id, ok := c.data[slug]
	return id, ok
}
