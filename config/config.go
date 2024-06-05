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

// 获取WithdrawWords
func GetWithdrawWords() []string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WithdrawWords
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

// 获取CheckVideoQRCode
func GetCheckVideoQRCode() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.CheckVideoQRCode
	}
	return false
}

// 获取QRLimit
func GetQRLimit() int {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.QRLimit
	}
	return 1
}

// 获取WithdrawNotice
func GetWithdrawNotice() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.WithdrawNotice
	}
	return ""
}

// 获取启用视频检查的配置
func GetOnEnableVideoCheck() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.OnEnableVideoCheck
	}
	return ""
}

// 获取禁用视频检查的配置
func GetOnDisableVideoCheck() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.OnDisableVideoCheck
	}
	return ""
}

// 获取启用图片检查的配置
func GetOnEnablePicCheck() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.OnEnablePicCheck
	}
	return ""
}

// 获取禁用图片检查的配置
func GetOnDisablePicCheck() string {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.OnDisablePicCheck
	}
	return ""
}

// GetSetGroupKick 获取 SetGroupKick 配置值
func GetSetGroupKick() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.SetGroupKick
	}
	return false
}

// GetKickAndRejectAddRequest 获取 KickAndRejectAddRequest 配置值
func GetKickAndRejectAddRequest() bool {
	mu.Lock()
	defer mu.Unlock()
	if instance != nil {
		return instance.Settings.KickAndRejectAddRequest
	}
	return false
}
