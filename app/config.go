package app

import "os"

type Config struct {
	Network string
}

var config *Config

func ReloadConfig() {
	if config == nil {
		config = &Config{}
	}
	config.Network = os.Getenv("NETWORK")
}

func Network() string {
	return config.Network
}

func IsTestnet() bool {
	return config.Network == "testnet"
}
