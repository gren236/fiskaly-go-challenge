package app

type Config struct {
	Host string `env:"API_HOST" validate:"required,ip4_addr"`
	Port int    `env:"API_PORT" validate:"gte=0,lte=65535"`
}

func NewConfig() Config {
	return Config{
		Host: "0.0.0.0",
		Port: 8080,
	}
}
