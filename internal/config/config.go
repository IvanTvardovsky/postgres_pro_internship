package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"sync"
)

type AppConfig struct {
	Directories []DirectoryConfig `yaml:"directories"`
	DataBase    DataBaseConfig    `yaml:"storage"`
}

type DirectoryConfig struct {
	Path          string   `yaml:"path"`
	Commands      []string `yaml:"commands"`
	IncludeRegexp []string `yaml:"include_regexp"`
	ExcludeRegexp []string `yaml:"exclude_regexp"`
	LogFile       string   `yaml:"log_file"`
}

type DataBaseConfig struct {
	Host     string `yaml:"host"`
	Port     rune   `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var instance *AppConfig
var once sync.Once

func GetConfig() *AppConfig {
	once.Do(func() {
		instance = &AppConfig{}
		if err := cleanenv.ReadConfig("config.yml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			log.Fatalf("Error reading config file: %v\n%v", err, help)
		}
	})
	return instance
}
