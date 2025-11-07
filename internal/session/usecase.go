package session

import (
	"context"
	"database/sql"
	stderr "errors"
	"time"

	"github.com/tuanta7/hydros/core/x"
	"github.com/tuanta7/hydros/internal/errors"
	"github.com/tuanta7/hydros/pkg/dbtype"
)

type UseCase interface {
	GetRememberedLoginSession(ctx context.Context, loginSessionFromCookie *LoginSession, id string) (*LoginSession, error)
	CreateLoginSession(ctx context.Context, session *LoginSession) error
	DeleteLoginSession(ctx context.Context, id string) (deletedSession *LoginSession, err error)
	RevokeSubjectLoginSession(ctx context.Context, user string) error
	ConfirmLoginSession(ctx context.Context, loginSession *LoginSession) error
}

type useCase struct {
	sessionRepo *Repository
}

func NewUseCase(sessionRepo *Repository) UseCase {
	return &useCase{
		sessionRepo: sessionRepo,
	}
}

func (u *useCase) GetRememberedLoginSession(ctx context.Context, loginSessionFromCookie *LoginSession, id string) (*LoginSession, error) {
	if loginSessionFromCookie != nil && loginSessionFromCookie.ID == id {
		return loginSessionFromCookie, nil
	}

	s, err := u.sessionRepo.GetRememberedLoginSession(ctx, id)
	if stderr.Is(err, sql.ErrNoRows) {
		return nil, errors.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return s, nil
}

func (u *useCase) CreateLoginSession(ctx context.Context, session *LoginSession) error {
	return nil
}

func (u *useCase) DeleteLoginSession(ctx context.Context, id string) (*LoginSession, error) {
	session, err := u.sessionRepo.DeleteLoginSession(ctx, id)
	if stderr.Is(err, sql.ErrNoRows) {
		return nil, errors.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return session, nil
}

func (u *useCase) RevokeSubjectLoginSession(ctx context.Context, user string) error {
	//TODO implement me
	panic("implement me")
}

func (u *useCase) ConfirmLoginSession(ctx context.Context, session *LoginSession) error {
	session.AuthenticatedAt = dbtype.NullTime(x.NowUTC().Truncate(time.Second))
	err := u.sessionRepo.UpsertLoginSession(ctx, session)
	if stderr.Is(err, sql.ErrNoRows) {
		return errors.ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}
