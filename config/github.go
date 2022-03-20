package config

// GithubHost represents a single host for which align has a configuration
type GithubHost struct {
	Token  string  `yaml:"token"`
	Limits *Limits `yaml:"limits"`
}

// Limits represents a limits override for the client
type Limits struct {
	RequestsPerSecond int `yaml:"request_per_second"`
	Burst             int `yaml:"burst"`
}
