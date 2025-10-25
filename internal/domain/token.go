package domain

import "database/sql"

type AccessToken struct {
	Signature string `db:"signature"`
}

type RefreshToken struct {
	Signature            string         `db:"signature"`
	AccessTokenSignature sql.NullString `db:"access_token_signature"`
}
