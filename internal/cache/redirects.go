package cache

import (
	"sync"
	"time"

	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-proxy/internal/repository"
	"go.uber.org/zap"
)

type RedirectEntry struct {
	TargetURL  string
	StatusCode int
}

type redirectKey struct {
	TenantID   string
	SourceSlug string
}

type RedirectCache struct {
	mu   sync.RWMutex
	data map[redirectKey]RedirectEntry
	repo repository.RedirectRepo
	last time.Time
}

func NewRedirectCache(repo repository.RedirectRepo) *RedirectCache {
	return &RedirectCache{
		data: make(map[redirectKey]RedirectEntry),
		repo: repo,
	}
}

func (c *RedirectCache) LoadFull() error {
	rows, err := c.repo.GetActiveRedirects()
	if err != nil {
		return err
	}

	data := make(map[redirectKey]RedirectEntry, len(rows))
	for _, r := range rows {
		data[redirectKey{TenantID: r.TenantID, SourceSlug: r.SourceSlug}] = RedirectEntry{
			TargetURL:  r.TargetURL,
			StatusCode: r.StatusCode,
		}
	}

	c.mu.Lock()
	c.data = data
	c.last = time.Now()
	c.mu.Unlock()

	logger.Get().Infow("redirect cache loaded", zap.Int("count", len(data)))
	return nil
}

func (c *RedirectCache) Sync() error {
	c.mu.RLock()
	since := c.last
	c.mu.RUnlock()

	rows, err := c.repo.GetUpdatedRedirects(since)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range rows {
		key := redirectKey{TenantID: r.TenantID, SourceSlug: r.SourceSlug}
		if r.IsActive {
			c.data[key] = RedirectEntry{TargetURL: r.TargetURL, StatusCode: r.StatusCode}
		} else {
			delete(c.data, key)
		}
	}

	c.last = time.Now()

	if len(rows) > 0 {
		logger.Get().Infow("redirect cache synced", zap.Int("updated", len(rows)))
	}
	return nil
}

func (c *RedirectCache) Lookup(tenantID, sourceSlug string) (RedirectEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.data[redirectKey{TenantID: tenantID, SourceSlug: sourceSlug}]
	return entry, ok
}
