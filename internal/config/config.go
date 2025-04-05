package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"prod"`
	StoragePath string        `yaml:"storage-path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"tokenTTL" env-default:"30m"`
	GRPC        GRPCObj       `yaml:"grpc"`
}

type GRPCObj struct {
	Port    int           `yaml:"port" env-default:"20202"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

const (
	defaultConfigPath = "./config/config.yml"
)

// New returns new config object
func New() *Config {
	path := fetchConfigPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file is not found")
	}

	var cfg Config
	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic("failed to parse config file" + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches the config file path either from flag "config" or
// environment variable "CONFIG_PATH". If no both of them, default path is returned
//
// flag > env > default
func fetchConfigPath() string {
	res := ""

	flag.StringVar(&res, "config", "", "path to the config file")
	flag.Parse()
	if res != "" {
		return res
	}

	res = os.Getenv("CONFIG_PATH")
	if res != "" {
		return res
	}

	res = defaultConfigPath
	return res
}
