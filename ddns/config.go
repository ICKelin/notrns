package ddns

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	DDNSConfig  *DDNSConfig  `json:"ddns" toml:"ddns"`
	StoreConfig *StoreConfig `json:"store" toml:"store"`
	ApiConfig   *ApiConfig   `json:"api" toml:"api"`
}

func ParseConfig(path string) (*Config, error) {
	cnt, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(cnt)
}

func parseConfig(cnt []byte) (*Config, error) {
	config := Config{}
	err := json.Unmarshal(cnt, &config)

	if config.DDNSConfig == nil {
		config.DDNSConfig = &DDNSConfig{}
	}

	if config.StoreConfig == nil {
		config.StoreConfig = &StoreConfig{}
	}

	if config.ApiConfig == nil {
		config.ApiConfig = &ApiConfig{}
	}

	return &config, err
}

func (c *Config) String() string {
	bytes, _ := json.Marshal(c)
	return string(bytes)
}
