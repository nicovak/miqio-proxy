package repository

import (
	"time"

	rp "github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-proxy/internal/model"
)

var _ DomainRepo = (*DomainRepository)(nil)

type DomainRepository struct {
	Repository rp.Repo
}

type DomainRepo interface {
	GetActiveDomains() ([]model.Domain, error)
	GetUpdatedDomains(since time.Time) ([]model.Domain, error)
}

func NewDomainRepository(repo rp.Repo) *DomainRepository {
	return &DomainRepository{Repository: repo}
}

func (r *DomainRepository) GetActiveDomains() ([]model.Domain, error) {
	var rows []model.Domain
	err := r.Repository.Select(
		`SELECT domain, tenant_id::text FROM tenant_domains WHERE status = 'ssl_active'`,
		&rows,
	)
	return rows, err
}

func (r *DomainRepository) GetUpdatedDomains(since time.Time) ([]model.Domain, error) {
	var rows []model.Domain
	err := r.Repository.Select(
		`SELECT domain, tenant_id::text, status FROM tenant_domains WHERE updated_at > $1`,
		&rows,
		since,
	)
	return rows, err
}
