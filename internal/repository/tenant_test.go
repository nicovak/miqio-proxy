package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	rp "github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-go/test"
	"github.com/nicovak/miqio-proxy/internal/model"
	"github.com/stretchr/testify/suite"
)

type TenantSuite struct {
	suite.Suite

	test.Postgres
	test.Fixture
}

func (s *TenantSuite) TestGetActiveTenants() {
	s.Create(&model.Tenant{
		ID:     "550e8400-e29b-41d4-a716-446655440001",
		Slug:   "acme",
		Status: "active",
	})
	s.Create(&model.Tenant{
		ID:     "550e8400-e29b-41d4-a716-446655440002",
		Slug:   "inactive-co",
		Status: "inactive",
	})

	repo := NewTenantRepository(&rp.PgxRepository{DB: test.GetPostgresConnPool()})
	result, err := repo.GetActiveTenants()

	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Require().Equal("acme", result[0].Slug)
	s.Require().Equal("550e8400-e29b-41d4-a716-446655440001", result[0].ID)
}

func (s *TenantSuite) TestGetActiveTenantsErr() {
	fakeDB, _ := pgxpool.New(context.Background(), "")
	repo := NewTenantRepository(&rp.PgxRepository{DB: fakeDB})
	result, err := repo.GetActiveTenants()

	s.Require().Error(err)
	s.Require().Empty(result)
}

func TestTenant(t *testing.T) {
	s := &TenantSuite{
		Fixture: test.FixtureChain{
			Fixtures: []test.Fixture{
				test.PostgresFixture{},
			},
		},
	}
	suite.Run(t, s)
}
