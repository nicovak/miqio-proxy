package model

import m "github.com/nicovak/miqio-go/model"

var _ m.Model = (*Domain)(nil)

type Domain struct {
	Domain   string `db:"domain"`
	TenantID string `db:"tenant_id"`
	Status   string `db:"status"`
}

func (mdl *Domain) TableName() string {
	return "tenant_domains"
}
