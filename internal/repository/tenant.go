package repository

import (
	"time"

	rp "github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-proxy/internal/model"
)

var _ TenantRepo = (*TenantRepository)(nil)

type TenantRepository struct {
	Repository rp.Repo
}

type TenantRepo interface {
	GetActiveTenants() ([]model.Tenant, error)
	GetUpdatedTenants(since time.Time) ([]model.Tenant, error)
}

func NewTenantRepository(repo rp.Repo) *TenantRepository {
	return &TenantRepository{Repository: repo}
}

func (r *TenantRepository) GetActiveTenants() ([]model.Tenant, error) {
	var rows []model.Tenant
	err := r.Repository.Select(
		`SELECT slug, id::text FROM tenants WHERE status = 'active'`,
		&rows,
	)
	return rows, err
}

func (r *TenantRepository) GetUpdatedTenants(since time.Time) ([]model.Tenant, error) {
	var rows []model.Tenant
	err := r.Repository.Select(
		`SELECT slug, id::text, status FROM tenants WHERE updated_at > $1`,
		&rows,
		since,
	)
	return rows, err
}
