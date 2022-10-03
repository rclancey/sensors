package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rclancey/argparse"
	"github.com/rclancey/httpserver/v2"
)

type Config struct {
	*httpserver.ServerConfig
	Network *httpserver.NetworkConfig `json:"network"`
	Location *Location `json:"location"`
	OpenWeatherMapAPIKey string `json:"openweathermap"`
	AppleTV *AppleTVCfg `json:"appletv"`
	Router *RouterCfg `json:"router"`
}

type AppleTVCfg struct {
	ID string `json:"id"`
	Airplay string `json:"airplay"`
	Companion string `json:"companion"`
}

type RouterCfg struct {
	IP string `json:"ip"`
	Password string `json:"password"`
	Hosts map[string]string `json:"hosts"`
}

func (cfg *Config) LoadFromFile(fn string) error {
	var f io.ReadCloser
	var err error
	if fn == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(fn)
		if err != nil {
			return err
		}
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, cfg)
}

func (cfg *Config) Init() error {
	err := cfg.ServerConfig.Init()
	if err != nil {
		return err
	}
	return nil
}

func DefaultConfig() *Config {
	return &Config{
		ServerConfig: httpserver.DefaultServerConfig(),
		Network: &httpserver.NetworkConfig{},
		Location: &Location{},
	}
}

func Configure() (*Config, error) {
	cfg := DefaultConfig()
	err := argparse.ParseArgs(cfg)
	if err != nil {
		return nil, err
	}
	cfg.ServerRoot, err = filepath.Abs(filepath.Clean(httpserver.EnvEval(cfg.ServerRoot)))
	if err != nil {
		return nil, err
	}
	cfg.ConfigFile, err = cfg.Abs(cfg.ConfigFile)
	if err != nil {
		return nil, err
	}
	_, err = os.Stat(cfg.ConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		err = cfg.LoadFromFile(cfg.ConfigFile)
		if err != nil {
			return nil, err
		}
		err = argparse.ParseArgs(cfg)
		if err != nil {
			return nil, err
		}
	}
	err = cfg.Init()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
