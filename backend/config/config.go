package config

type Config struct {
	Chain  Chain  `yaml:"chain"`
	Faucet Faucet `yaml:"faucet"`
	Log    Log    `yaml:"log"`
	Redis  Redis  `yaml:"redis"`
	RPC    RPC    `yaml:"rpc"`
	Server Server `yaml:"server"`
	Wallet Wallet `yaml:"wallet"`
}
