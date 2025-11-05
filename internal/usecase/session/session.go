package session

import (
	"context"
	"database/sql"
	"errors"

	"github.com/tuanta7/hydros/internal/datasource/postgres"
	"github.com/tuanta7/hydros/internal/domain"
)

type UseCase interface {
	GetRememberedLoginSession(ctx context.Context, loginSessionFromCookie *domain.LoginSession, id string) (*domain.LoginSession, error)
	CreateLoginSession(ctx context.Context, session *domain.LoginSession) error
	DeleteLoginSession(ctx context.Context, id string) (deletedSession *domain.LoginSession, err error)
	RevokeSubjectLoginSession(ctx context.Context, user string) error
	ConfirmLoginSession(ctx context.Context, loginSession *domain.LoginSession) error
}

type useCase struct {
	sessionRepo *postgres.SessionRepository
}

func NewUseCase(sessionRepo *postgres.SessionRepository) UseCase {
	return &useCase{
		sessionRepo: sessionRepo,
	}
}

func (u *useCase) GetRememberedLoginSession(ctx context.Context, loginSessionFromCookie *domain.LoginSession, id string) (*domain.LoginSession, error) {
	if loginSessionFromCookie != nil && loginSessionFromCookie.ID == id {
		return loginSessionFromCookie, nil
	}

	s, err := u.sessionRepo.GetRememberedLoginSession(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		// TODO: read hydra errors implementation
		return nil, domain.ErrNoAuthenticationSessionFound
	} else if err != nil {
		return nil, err
	}

	return s, nil
}

func (u *useCase) CreateLoginSession(ctx context.Context, session *domain.LoginSession) error {
	//TODO implement me
	panic("implement me")
}

func (u *useCase) DeleteLoginSession(ctx context.Context, id string) (deletedSession *domain.LoginSession, err error) {
	//TODO implement me
	panic("implement me")
}

func (u *useCase) RevokeSubjectLoginSession(ctx context.Context, user string) error {
	//TODO implement me
	panic("implement me")
}

func (u *useCase) ConfirmLoginSession(ctx context.Context, loginSession *domain.LoginSession) error {
	//TODO implement me
	panic("implement me")
}
