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

type RedirectSuite struct {
	suite.Suite

	test.Postgres
	test.Fixture
}

func (s *RedirectSuite) TestGetActiveRedirects() {
	s.Create(&model.Redirect{
		TenantID:   "550e8400-e29b-41d4-a716-446655440001",
		SourceSlug: "old-page",
		TargetURL:  "https://example.com/new-page",
		StatusCode: 301,
		IsActive:   true,
	})
	s.Create(&model.Redirect{
		TenantID:   "550e8400-e29b-41d4-a716-446655440001",
		SourceSlug: "disabled",
		TargetURL:  "https://example.com/other",
		StatusCode: 302,
		IsActive:   false,
	})

	repo := NewRedirectRepository(&rp.PgxRepository{DB: test.GetPostgresConnPool()})
	result, err := repo.GetActiveRedirects()

	s.Require().NoError(err)
	s.Require().Len(result, 1)
	s.Require().Equal("old-page", result[0].SourceSlug)
	s.Require().Equal("https://example.com/new-page", result[0].TargetURL)
	s.Require().Equal(301, result[0].StatusCode)
}

func (s *RedirectSuite) TestGetActiveRedirectsErr() {
	fakeDB, _ := pgxpool.New(context.Background(), "")
	repo := NewRedirectRepository(&rp.PgxRepository{DB: fakeDB})
	result, err := repo.GetActiveRedirects()

	s.Require().Error(err)
	s.Require().Empty(result)
}

func TestRedirect(t *testing.T) {
	s := &RedirectSuite{
		Fixture: test.FixtureChain{
			Fixtures: []test.Fixture{
				test.PostgresFixture{},
			},
		},
	}
	suite.Run(t, s)
}
