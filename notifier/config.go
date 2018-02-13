package notifier

// MailConfig of smtp sender
type MailConfig struct {
	From        string `yaml:"from"`
	SMTPHost    string `yaml:"host"`
	SMTPPort    int    `yaml:"port"`
	InsecureTLS bool   `yaml:"insecure_tls"`
}

type Sender interface{}

// Config setting
type Config struct {
	ElasticUrls []string `yaml:"elasticsearch_url"`
	Senders     []Sender `yaml:"senders"`
	StatsD      string   `yaml:"statsd"`
}
