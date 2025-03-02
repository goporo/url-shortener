package config

type Config struct {
	RateLimit struct {
		Enabled           bool
		RequestsPerMinute int
	}
}

func GetDefaultConfig() *Config {
	config := &Config{}

	config.RateLimit.Enabled = true
	config.RateLimit.RequestsPerMinute = 60

	return config
}
