package main

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type f5Config struct {
	AuthMethod        string `toml:"auth_method"`
	URL               string `toml:"url"`
	User              string `toml:"user"`
	Password          string `toml:"password"`
	SSLCheck          bool   `toml:"ssl_check"`
	LoginProviderName string `toml:"login_provided_name"`
}

type watchConfig struct {
	Dir string `toml:"directory"`
}

type config struct {
	F5    f5Config    `toml:"f5"`
	Watch watchConfig `toml:"watch"`
}

func readConfig(path string) (*config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("cannot open configuration file: " + err.Error())
	}
	defer file.Close()

	var cfg config
	if _, err := toml.DecodeReader(file, &cfg); err != nil {
		return nil, errors.New("cannot read configuration file: " + err.Error())
	}

	return &cfg, nil
}
