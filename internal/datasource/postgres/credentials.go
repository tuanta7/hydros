package postgres

import "time"

type Credentials struct {
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	PIN       string    `json:"pin" db:"pin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
