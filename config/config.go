package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Version        string `koanf:"version"`
	RestServerHost string `koanf:"rest_server_host"`
	RestServerPort string `koanf:"rest_server_port"`
	GrpcServerHost string `koanf:"grpc_server_host"`
	GrpcServerPort string `koanf:"grpc_server_port"`

	AuthorizationCodeLifetime time.Duration `koanf:"authorization_code_lifetime" default:"10m"`
}

func LoadEnvConfig(envFiles ...string) *Config {
	err := godotenv.Load(envFiles...)
	if err != nil {
		log.Printf("[Warning] config - init - godotenv.Load: %v", err)
	}

	cfg := &Config{}
	err = envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("config - init - envconfig.Process: %v", err)
	}
	return cfg
}

func (c *Config) LoadJSONConfig(k *koanf.Koanf) {
	f := file.Provider("config/config.json")
	if err := k.Load(f, json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if ue := k.Unmarshal("", c); ue != nil {
		log.Printf("error unmarshalling config: %v", ue)
		return
	}

	go func() {
		err := f.Watch(func(event any, err error) {
			if err != nil {
				log.Printf("watch error: %v", err)
				return
			}

			log.Println("config changed. Reloading ...")
			if le := k.Load(f, json.Parser()); le != nil {
				log.Printf("error loading config: %v", le)
				return
			}

			if ue := k.Unmarshal("", c); ue != nil {
				log.Printf("error unmarshalling config: %v", ue)
				return
			}

			k.Print()
		})
		if err != nil {
			return
		}
	}()
}
