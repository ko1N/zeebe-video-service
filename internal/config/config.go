package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// TODO: allow configuration of video2x, rife, ffmpeg binary locations here!

type Config struct {
	Zeebe ZeebeConfig `yaml:"zeebe"`
}

type ZeebeConfig struct {
	Host      string `yaml:"host"`
	Plaintext bool   `yaml:"plaintext"`
}

func ReadConfig(filename string) (*Config, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = yaml.Unmarshal([]byte(contents), &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
