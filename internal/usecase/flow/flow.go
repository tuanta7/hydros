package flow

type UseCase struct {
}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func (u *UseCase) UpdateAuthenticationStatus() {}

func (u *UseCase) UpdateConsentStatus() {}
