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
	LoginProviderName string `toml:"login_provider_name"`
}

type watchConfig struct {
	Dir               string   `toml:"directory"`
	Exclude           []string `toml:"exclude"`
	RemoveRemoveFiles bool     `toml:"remove_remote_files"`
}

type config struct {
	F5 f5Config `toml:"f5"`

	CredentialStorage string `toml:"credential_storage"` // "plain" or "secret"
	SecretStorePath   string `toml:"secret_store_path"`  // when CredentialStorage is "secret"
	Passphrase        string `toml:"token"`              // when CredentialStorage is "secret"

	Watch []watchConfig `toml:"watch"`
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
