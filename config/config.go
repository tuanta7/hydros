package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Version        string `koanf:"version"`
	RestServerHost string `koanf:"rest_server_host"`
	RestServerPort string `koanf:"rest_server_port"`
	GrpcServerHost string `koanf:"grpc_server_host"`
	GrpcServerPort string `koanf:"grpc_server_port"`
	GlobalSecret   string `koanf:"global_secret"`
	KeyEntropy     int    `koanf:"key_entropy"`

	Lifetime    *LifetimeConfig    `koanf:"lifetime"`
	Obfuscation *ObfuscationConfig `koanf:"obfuscation"`
	Redis       *RedisConfig       `koanf:"redis"`
	Postgres    *PostgresConfig    `koanf:"postgres"`
}

type RedisConfig struct {
	Host     string `koanf:"host"`
	Port     uint16 `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type PostgresConfig struct {
	Host     string `koanf:"host"`
	Port     uint16 `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	Database string `koanf:"database"`
	Params   map[string]string
}

func LoadConfig(envFiles ...string) *Config {
	err := godotenv.Load(envFiles...)
	if err != nil {
		log.Fatalf("No .env file found or error loading .env file: %v", err)
	}

	k := koanf.New(".")
	err = k.Load(env.Provider("HYDROS", ".", func(s string) string {
		s = strings.TrimPrefix(s, "HYDROS.")
		s = strings.ToLower(s)
		return s
	}), nil)
	if err != nil {
		log.Fatalf("error loading env config: %v", err)
	}

	f := file.Provider("config/config.json")
	err = k.Load(f, json.Parser())
	if err != nil {
		log.Fatalf("error loading json config: %v", err)
	}

	cfg := &Config{}
	if ue := k.Unmarshal("", cfg); ue != nil {
		log.Fatalf("error unmarshalling config: %v", ue)
	}

	go func() {
		we := f.Watch(func(event any, err error) {
			if err != nil {
				log.Printf("watch error: %v", err)
				return
			}

			log.Println("config changed. Reloading ...")
			if le := k.Load(f, json.Parser()); le != nil {
				log.Printf("error loading config: %v", le)
				return
			}

			if ue := k.Unmarshal("", cfg); ue != nil {
				log.Printf("error unmarshalling config: %v", ue)
				return
			}

			k.Print()
		})
		if we != nil {
			return
		}
	}()

	return cfg
}
