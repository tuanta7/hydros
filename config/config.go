package config

import (
	"log"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	*koanf.Koanf

	Version string `koanf:"version"`
}

func LoadJSONConfig(c *Config) {
	var k = koanf.New(".")

	f := file.Provider("config/config.json")
	if err := k.Load(f, json.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	c.Koanf = k

	go func() {
		err := f.Watch(func(event any, err error) {
			if err != nil {
				log.Printf("watch error: %v", err)
				return
			}

			log.Println("config changed. Reloading ...")
			if err := c.Koanf.Load(f, json.Parser()); err != nil {
				log.Printf("error loading config: %v", err)
				return
			}

			if err := c.Koanf.Unmarshal("", c); err != nil {
				log.Printf("error unmarshalling config: %v", err)
				return
			}

			k.Print()
		})
		if err != nil {
			return
		}
	}()
}
