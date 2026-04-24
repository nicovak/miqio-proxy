package cache

import (
	"sync"
	"time"

	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-proxy/internal/repository"
	"go.uber.org/zap"
)

type DomainCache struct {
	mu   sync.RWMutex
	data map[string]string // domain -> tenantID
	repo repository.DomainRepo
	last time.Time
}

func NewDomainCache(repo repository.DomainRepo) *DomainCache {
	return &DomainCache{
		data: make(map[string]string),
		repo: repo,
	}
}

func (c *DomainCache) LoadFull() error {
	rows, err := c.repo.GetActiveDomains()
	if err != nil {
		return err
	}

	data := make(map[string]string, len(rows))
	for _, r := range rows {
		data[r.Domain] = r.TenantID
	}

	c.mu.Lock()
	c.data = data
	c.last = time.Now()
	c.mu.Unlock()

	logger.Get().Infow("domain cache loaded", zap.Int("count", len(data)))
	return nil
}

func (c *DomainCache) Sync() error {
	c.mu.RLock()
	since := c.last
	c.mu.RUnlock()

	rows, err := c.repo.GetUpdatedDomains(since)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range rows {
		if r.Status == "ssl_active" {
			c.data[r.Domain] = r.TenantID
		} else {
			delete(c.data, r.Domain)
		}
	}

	c.last = time.Now()

	if len(rows) > 0 {
		logger.Get().Infow("domain cache synced", zap.Int("updated", len(rows)))
	}
	return nil
}

func (c *DomainCache) Resolve(domain string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	id, ok := c.data[domain]
	return id, ok
}
