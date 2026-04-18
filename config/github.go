package config

// GithubHost holds authentication and rate-limit settings for a GitHub host.
type GithubHost struct {
	Token          string  `yaml:"token"`
	Username       string  `yaml:"username"`
	PrivateKey     string  `yaml:"private_key"`
	PrivateKeyFile string  `yaml:"private_key_file"`
	Limits         *Limits `yaml:"limits"`
}

// Limits configures GitHub API client request rates.
type Limits struct {
	RequestsPerSecond int `yaml:"request_per_second"`
	Burst             int `yaml:"burst"`
}
