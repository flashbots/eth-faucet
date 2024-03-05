package config

import "time"

type Redis struct {
	Namespace string        `yaml:"namespace"`
	Timeout   time.Duration `yaml:"timeout"`
	URL       string        `yaml:"url"`
}
