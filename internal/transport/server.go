package transport

import "context"

type Server interface {
	Run() error
	Shutdown(context.Context) error
}
