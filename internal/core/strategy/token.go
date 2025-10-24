package strategy

// TokenStrategy defines the methods used for managing 
// Access Token, Refresh Token, or Authorization Code
type TokenStrategy interface {
	Generate(ctx context.Context) (token string, signature string, err error)
	GetSignature(ctx context.Context, token string) string
	Validate(ctx context.Context, request core.Request, token string) (err error)
}