package login

type UseCase struct {
}

func (u *UseCase) RequestAuthentication() {}

func (u *UseCase) UpdateAuthenticationStatus() {}

func (u *UseCase) VerifyAuthentication() {}

func (u *UseCase) RequestConsent() {}

func (u *UseCase) UpdateConsentStatus() {}

func (u *UseCase) VerifyConsent() {}
