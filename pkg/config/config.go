package config

import (
	"os"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	AppID      int64
	PrivateKey []byte
)

type Config struct {
	Server HTTPConfig       `yaml:"server"`
	Github githubapp.Config `yaml:"github"`
	DB     DBConfig         `yaml:"database"`
	Log    LogConfig        `yaml:"log"`

	AppConfig CrownConfig `yaml:"app_configuration"`
}

// DBConfig
type DBConfig struct {
	Path string `yaml:"path"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	Human bool   `yaml:"human"`
}

type HTTPConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type CrownConfig struct {
	PullRequestPreamble string `yaml:"pull_request_preamble"`
}

func ReadConfig(path string) (*Config, error) {
	var c Config

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading server config file: %s", path)
	}

	if err := yaml.Unmarshal(bytes, &c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}

	if c.Log.Level == "" {
		c.Log.Level = "info"
	}

	AppID = c.Github.App.IntegrationID
	PrivateKey = []byte(c.Github.App.PrivateKey)

	return &c, nil
}
