package elastic

// Config setting
type Config struct {
	Version             string   `yaml:"version"`
	ElasticUrls         []string `yaml:"elasticsearch_url"`
	BulkSize            uint     `yaml:"bulk_size"`
	BulkRefreshInterval int64    `yaml:"bulk_refresh_interval"`
	ConcurentWrites     uint     `yaml:"concurent_writes"`
}
