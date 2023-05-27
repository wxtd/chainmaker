package config

type userConfig struct {
	Type       string         `mapstructure:"type"`
	Count      int32          `mapstructure:"count"`
	Location   locationConfig `mapstructure:"location"`
	ExpireYear int32          `mapstructure:"expire_year"`
}

type nodeConfig struct {
	Type     string         `mapstructure:"type"`
	Count    int32          `mapstructure:"count"`
	Location locationConfig `mapstructure:"location"`
	Specs    specsConfig    `mapstructure:"specs"`
}

type specsConfig struct {
	ExpireYear int32    `mapstructure:"expire_year"`
	SANS       []string `mapstructure:"sans"`
}

type caConfig struct {
	Location locationConfig `mapstructure:"location"`
	Specs    specsConfig    `mapstructure:"specs"`
}

type itemConfig struct {
	Domain   string         `mapstructure:"domain"`
	HostName string         `mapstructure:"host_name"`
	PKAlgo   string         `mapstructure:"pk_algo"`
	SKIHash  string         `mapstructure:"ski_hash"`
	Specs    specsConfig    `mapstructure:"specs"`
	Location locationConfig `mapstructure:"location"`
	Count    int32          `mapstructure:"count"`
	CA       caConfig       `mapstructure:"ca"`
	Node     []nodeConfig   `mapstructure:"node"`
	User     []userConfig   `mapstructure:"user"`
}

type locationConfig struct {
	Country  string `mapstructure:"country"`
	Locality string `mapstructure:"locality"`
	Province string `mapstructure:"province"`
}

type CryptoGenConfig struct {
	Item []itemConfig `mapstructure:"crypto_config"`
}
