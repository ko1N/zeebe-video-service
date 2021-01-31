package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Zeebe   ZeebeConfig    `yaml:"zeebe"`
	FFmpeg  *FFmpegConfig  `yaml:"ffmpeg"`
	Video2x *Video2xConfig `yaml:"video2x"`
	Rife    *RifeConfig    `yaml:"rife"`
}

type ZeebeConfig struct {
	Host      string `yaml:"host"`
	Plaintext bool   `yaml:"plaintext"`
}

type FFmpegConfig struct {
	FFprobeExecutable string `yaml:"ffprobe"`
	FFmpegExecutable  string `yaml:"ffmpeg"`
}

type Video2xConfig struct {
	Executable string `yaml:"executable"`
}

type RifeConfig struct {
	Executable string `yaml:"executable"`
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
