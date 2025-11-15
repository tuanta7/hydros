package login

type Credentials struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

func (c *Credentials) ColumnMap() map[string]any {
	return map[string]any{
		"username": c.Username,
		"password": c.Password,
	}
}
