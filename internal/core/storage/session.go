package storage

type AuthorizeCodeSession interface {}

type AccessTokenSession interface {}

type RefreshTokenSession interface {
	DeleteSession(ctx context.Context, signature string) error
}