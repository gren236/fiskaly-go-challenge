package app

type Config struct {
	ApiHost string `env:"API_HOST" validate:"required,ip4_addr"`
	ApiPort int    `env:"API_PORT" validate:"gte=0,lte=65535"`
}

func NewConfig() Config {
	return Config{
		ApiHost: "0.0.0.0",
		ApiPort: 8080,
	}
}
