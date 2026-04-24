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

type DomainSuite struct {
	suite.Suite

	test.Postgres
	test.Fixture
}

func (s *DomainSuite) TestGetActiveDomains() {
	s.Create(&model.Domain{
		Domain:   "example.com",
		TenantID: "550e8400-e29b-41d4-a716-446655440001",
		Status:   "ssl_active",
	})
	s.Create(&model.Domain{
		Domain:   "pending.com",
		TenantID: "550e8400-e29b-41d4-a716-446655440002",
		Status:   "pending",
	})

	repo := NewDomainRepository(&rp.PgxRepository{DB: test.GetPostgresConnPool()})
	result, err := repo.GetActiveDomains()

	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Require().Equal("example.com", result[0].Domain)
	s.Require().Equal("550e8400-e29b-41d4-a716-446655440001", result[0].TenantID)
}

func (s *DomainSuite) TestGetActiveDomainsErr() {
	fakeDB, _ := pgxpool.New(context.Background(), "")
	repo := NewDomainRepository(&rp.PgxRepository{DB: fakeDB})
	result, err := repo.GetActiveDomains()

	s.Require().Error(err)
	s.Require().Empty(result)
}

func TestDomain(t *testing.T) {
	s := &DomainSuite{
		Fixture: test.FixtureChain{
			Fixtures: []test.Fixture{
				test.PostgresFixture{},
			},
		},
	}
	suite.Run(t, s)
}
