package config

// GithubHost represents a single host for which align has a configuration
type GithubHost struct {
	Token          string  `yaml:"token"`
	Username       string  `yaml:"username"`
	PrivateKey     string  `yaml:"private_key"`
	PrivateKeyFile string  `yaml:"private_key_file"`
	Limits         *Limits `yaml:"limits"`
}

// Limits represents a limits override for the client
type Limits struct {
	RequestsPerSecond int `yaml:"request_per_second"`
	Burst             int `yaml:"burst"`
}
