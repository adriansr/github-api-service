package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/adriansr/github-api-service/util"
)

type Duration struct {
	Duration time.Duration
}

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

func (d *Duration) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		unquoted := string(b[1 : len(b)-1])
		parsed, err := time.ParseDuration(unquoted)
		if err != nil {
			return err
		}
		d.Duration = parsed
		return nil
	}
	return util.NewError("expected a string to decode a Duration")
}
