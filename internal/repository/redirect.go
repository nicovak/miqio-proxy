package repository

import (
	"time"

	rp "github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-proxy/internal/model"
)

var _ RedirectRepo = (*RedirectRepository)(nil)

type RedirectRepository struct {
	Repository rp.Repo
}

type RedirectRepo interface {
	GetActiveRedirects() ([]model.Redirect, error)
	GetUpdatedRedirects(since time.Time) ([]model.Redirect, error)
}

func NewRedirectRepository(repo rp.Repo) *RedirectRepository {
	return &RedirectRepository{Repository: repo}
}

func (r *RedirectRepository) GetActiveRedirects() ([]model.Redirect, error) {
	var rows []model.Redirect
	err := r.Repository.Select(
		`SELECT tenant_id::text, source_slug, target_url, status_code
		 FROM page_redirects WHERE is_active = true`,
		&rows,
	)
	return rows, err
}

func (r *RedirectRepository) GetUpdatedRedirects(since time.Time) ([]model.Redirect, error) {
	var rows []model.Redirect
	err := r.Repository.Select(
		`SELECT tenant_id::text, source_slug, target_url, status_code, is_active
		 FROM page_redirects WHERE updated_at > $1`,
		&rows,
		since,
	)
	return rows, err
}
