package postgres_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/suite"
	"github.com/tuanta7/hydros/core"
	"github.com/tuanta7/hydros/internal/domain"
)

type ClientRepositoryTestSuite struct {
	suite.Suite
}

func (s *ClientRepositoryTestSuite) SetupSuite() {}

func (s *ClientRepositoryTestSuite) TearDownSuite() {}

func (s *ClientRepositoryTestSuite) TestClientCreate() {
	client := &domain.Client{
		ID:          gofakeit.UUID(),
		Name:        gofakeit.Username(),
		Description: gofakeit.Comment(),

		TokenEndpointAuthMethod:     core.ClientAuthenticationMethodNone,
		TokenEndpointAuthSigningAlg: "none",
	}

	assert.Equal(s.T(), client.TokenEndpointAuthMethod, "none")
}

func TestClientRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ClientRepositoryTestSuite))
}
