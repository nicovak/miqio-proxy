package model

import m "github.com/nicovak/miqio-go/model"

var _ m.Model = (*Tenant)(nil)

type Tenant struct {
	Slug   string `db:"slug"`
	ID     string `db:"id"`
	Status string `db:"status"`
}

func (mdl *Tenant) TableName() string {
	return "tenants"
}
