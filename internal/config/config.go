package config

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Version        string `koanf:"version"`
	LogLevel       string `koanf:"log_level"`
	ReleaseMode    string `koanf:"release_mode"`
	RestServerHost string `koanf:"rest_server_host"`
	RestServerPort string `koanf:"rest_server_port"`
	GRPCServerHost string `koanf:"grpc_server_host"`
	GRPCServerPort string `koanf:"grpc_server_port"`

	Redis         RedisConfig         `koanf:"redis"`
	Postgres      PostgresConfig      `koanf:"postgres"`
	SessionCookie SessionCookieConfig `koanf:"session_cookie"`
	OAuth         OAuthConfig         `koanf:"oauth"`
	OIDC          OIDCConfig          `koanf:"oidc"`
	Lifetime      LifetimeConfig      `koanf:"lifetime"`
	HMAC          HMACConfig          `koanf:"hmac"`
	JWT           JWTConfig           `koanf:"jwt"`
	Obfuscation   ObfuscationConfig   `koanf:"obfuscation"`
	Identity      IdentityConfig      `koanf:"identity"`
}

func (c *Config) IsDebugging() bool {
	return c.ReleaseMode == "debug"
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

func (c *PostgresConfig) DSN(opts ...map[string]string) string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)

	if len(opts) > 0 {
		for k, v := range opts[0] {
			dsn += fmt.Sprintf("&%s=%s", k, v)
		}
	}

	dsn = strings.Replace(dsn, "&", "?", 1)
	return dsn
}

func LoadConfig(envFiles ...string) *Config {
	err := godotenv.Load(envFiles...)
	if err != nil {
		log.Fatalf("no .env file found or error loading .env file: %v", err)
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

	// JSON config will override env config
	f := file.Provider("static/config/config.json")
	if err = k.Load(f, json.Parser()); err != nil {
		log.Printf("error loading json config: %v", err)
		return
	}

	cfg := &Config{}
	if err = k.Unmarshal("", cfg); err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	if err = validateConfig(cfg); err != nil {
		log.Fatalf("invalid config values:\n%v", err)
	}

	go func() {
		we := f.Watch(func(event any, err error) {
			if err != nil {
				log.Printf("watch error: %v", err)
				return
			}

			log.Println("config changed, reloading ...")
			if err = k.Load(f, json.Parser()); err != nil {
				log.Printf("error loading config: %v", err)
				return
			}

			if err = k.Unmarshal("", cfg); err != nil {
				log.Printf("error unmarshalling config: %v", err)
				return
			}

			if err = validateConfig(cfg); err != nil {
				log.Printf("⚠️ invalid config values:\n%v", err)
			}

			k.Print()
		})
		if we != nil {
			return
		}
	}()

	return cfg
}

type ValidationError struct {
	Field string
	Tag   string
	Value any
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func validateConfig(cfg *Config) (validateErr error) {
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := f.Tag.Get("koanf")
		if name == "-" {
			return ""
		}

		return strings.ToUpper(name)
	})

	if errs := validate.Struct(cfg); errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validateErr = errors.Join(validateErr, fmt.Errorf(
				"invalid value for %s; constraints: %s, got `%s`",
				err.Field(),
				err.Tag(),
				err.Value(),
			))
		}
	}

	return validateErr
}
