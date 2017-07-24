// Package config contains a basic implementation of reading configuration
// from a json file
package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/adriansr/github-api-service/util"
)

type Config struct {
	Credentials GitHubCredentials `json:"github_credentials"`
	Client      HTTPClientConfig  `json:"client"`
	Server      HTTPServerConfig  `json:"server"`
}

type GitHubCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HTTPClientConfig struct {
	RequestTimeout Duration `json:"timeout"`
	ApiUrl         string   `json:"api_url"`
}

type HTTPServerConfig struct {
	ListenAddress string `json:"listen"`
}

func LoadRaw(content []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, util.WrapError("failed to parse configuration", err)
	}
	return &config, nil
}

func LoadFile(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, util.WrapError(
			"failed reading configuration file `"+path+"`", err)
	}
	return LoadRaw(content)
}
