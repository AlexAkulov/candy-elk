package http

// Config settings
type Config struct {
	Address     string              `yaml:"address"`
	APIKeys     map[string][]string `yaml:"api_keys"`
	Timeout     int64               `yaml:"timeout"`
	IdleTimeout int64               `yaml:"idle_timeout"`
}
