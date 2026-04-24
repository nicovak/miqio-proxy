package model

import m "github.com/nicovak/miqio-go/model"

var _ m.Model = (*Redirect)(nil)

type Redirect struct {
	TenantID   string `db:"tenant_id"`
	SourceSlug string `db:"source_slug"`
	TargetURL  string `db:"target_url"`
	StatusCode int    `db:"status_code"`
	IsActive   bool   `db:"is_active"`
}

func (mdl *Redirect) TableName() string {
	return "page_redirects"
}
