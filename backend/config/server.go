package config

type Server struct {
	AuthSecret         string `yaml:"auth_secret"`
	ListenAddress      string `yaml:"listen_address"`
	MaxRequestBodySize int    `yaml:"max_request_body_size"`
	ProxyCount         int    `yaml:"proxy_count"`
	RedisAddress       string `yaml:"redis_address"`
}
