package redis

import (
	goredis "github.com/redis/go-redis/v9"
)

type Option func(o *goredis.Options)

func WithCredential(username, password string) Option {
	return func(o *goredis.Options) {
		o.Username = username
		o.Password = password
	}
}

func WithDB(db int) Option {
	return func(o *goredis.Options) {
		o.DB = db
	}
}
