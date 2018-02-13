package metrics

// Config is settings for graphite
type Config struct {
	Enabled                  bool   `yaml:"enabled"`
	GraphiteConnectionString string `yaml:"graphite_connection_string"`
	GraphitePrefix           string `yaml:"graphite_prefix"`
}
