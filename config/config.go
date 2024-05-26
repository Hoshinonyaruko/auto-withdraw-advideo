package config

import (
	"log"
	"os"
	"sync"

	"github.com/hoshinonyaruko/auto-withdraw-advideo/structs"
	"gopkg.in/yaml.v3"
)

var (
	instance *Config
	mu       sync.Mutex
)

type Config struct {
	Version  int              `yaml:"version"`
	Settings structs.Settings `yaml:"settings"`
}

func LoadConfig(path string) (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	conf, err := loadConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	instance = conf
	return instance, nil
}

func loadConfigFromFile(path string) (*Config, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		log.Println("Failed to read file:", err)
		return nil, err
	}

	conf := &Config{}
	if err := yaml.Unmarshal(configData, conf); err != nil {
		log.Printf("failed to unmarshal YAML[%v]:%v", path, err)
		return nil, err
	}

	log.Printf("成功加载配置文件 %s\n", path)
	return conf, nil
}

// 获取HttpPaths
func GetHttpPaths() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPaths
	}
	return nil
}

// 获取VideoSecondLimit
func GetVideoSecondLimit() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.VideoSecondLimit
	}
	return 5
}

// 获取HttpPathsAccessTokens
func GetHttpPathsAccessTokens() []structs.AccessToken {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.HttpPathsAccessTokens
	}
	return nil
}
