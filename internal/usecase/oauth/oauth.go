package oauth

import (
	"context"
	"errors"

	"github.com/tuanta7/hydros/core"
)

type UseCase struct {
	authorizeInteractors  []core.AuthorizeInteractor
	tokenInteractors      []core.TokenInteractor
	accessTokenRepository AccessTokenRepository
}

func (u *UseCase) HandleTokenEndpoint(ctx context.Context, req *core.TokenRequest, res *core.TokenResponse) error {
	for _, ah := range u.tokenInteractors {
		err := ah.HandleTokenRequest(ctx, req, res)
		if errors.Is(err, core.ErrUnknownRequest) {
			// skip to next token interactor
			continue
		} else if err != nil {
			return err
		}
	}

	err := u.accessTokenRepository.Create(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
