package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	QBittorrentURL   string `yaml:"qbittorrent_url"`
	QBittorrentUser  string `yaml:"qbittorrent_user"`
	QBittorrentPass  string `yaml:"qbittorrent_password"`
	DownloadLocation string `yaml:"download_dir"`
	SpaceMargin      int    `yaml:"space_margin"`
	AutoResume       bool   `yaml:"autoresume"`
	LogFilePath      string `yaml:"log_file_path"`
	SkipForceResume  bool   `yaml:"skip_force_resume"`
}

func Load(configFilePath string) (Config, error) {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return Config{}, err
	}

	// Unmarshal the YAML data into a struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
