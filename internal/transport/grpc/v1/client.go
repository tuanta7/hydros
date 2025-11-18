package v1

import (
	"github.com/tuanta7/hydros/internal/config"
	"github.com/tuanta7/hydros/proto/gobuf/v1"
)

type ClientService struct {
	cfg *config.Config
	v1.UnimplementedClientServiceServer
}
