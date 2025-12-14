package postgres

import "github.com/jackc/pgx/v5"

func ToObject[T any](row pgx.CollectableRow) (*T, error) {
	c, err := pgx.RowToStructByName[T](row)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
