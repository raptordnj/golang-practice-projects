package classes

type Config struct {
	Host    string
	Port    int
	Timeout int
}

func NewConfig(host string) *Config {
	return &Config{
		Host:    host,
		Port:    8080,
		Timeout: 30,
	}
}
