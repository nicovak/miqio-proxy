package config

import (
	"time"

	cfg "github.com/nicovak/miqio-go/config"
)

type R2 struct {
	AccountID       string `env:"ACCOUNT_ID,notEmpty"`
	AccessKeyID     string `env:"ACCESS_KEY_ID,notEmpty"`
	SecretAccessKey  string `env:"SECRET_ACCESS_KEY,notEmpty"`
	BucketName      string `env:"BUCKET_NAME" envDefault:"miqio-lp-assets"`
}

type Config struct {
	Env               string        `env:"ENV"`
	Port              string        `env:"PORT" envDefault:"8080"`
	Postgres          cfg.Postgres  `envPrefix:"POSTGRES_"`
	R2                R2            `envPrefix:"R2_"`
	MiqioPageDomain   string        `env:"MIQIO_PAGE_DOMAIN" envDefault:"miqio.page"`
	CacheSyncInterval time.Duration `env:"CACHE_SYNC_INTERVAL" envDefault:"30s"`
	LogLevel          string        `env:"LOG_LEVEL" envDefault:"info"`
	AppOrigin         string        `env:"APP_ORIGIN" envDefault:"https://app.miqio.page"`
}

func Get() *Config {
	if conf, ok := cfg.Get().(*Config); ok {
		return conf
	}

	return &Config{}
}
