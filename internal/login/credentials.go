package login

type Credentials struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Extra    map[string]string `json:"extra"`
}

func (c *Credentials) ColumnMap() map[string]any {
	return map[string]any{
		"username": c.Username,
		"password": c.Password,
	}
}
