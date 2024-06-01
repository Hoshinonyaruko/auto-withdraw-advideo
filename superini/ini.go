package superini

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/ini.v1"
)

// ConfigManager manages the configuration settings stored in an INI file.
type ConfigManager struct {
	filePath       string
	data           *ini.File
	mu             sync.RWMutex
	isManualUpdate bool
}

var instance *ConfigManager
var once sync.Once

// Initialize the singleton instance.
func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{
			filePath: "config.ini",
		}
		instance.loadConfig()
	})
	return instance
}

// loadConfig loads the INI configuration from file.
func (m *ConfigManager) loadConfig() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		// 如果文件不存在，创建一个新的配置文件
		m.data = ini.Empty()
		// 可以在这里添加默认配置
		//m.data.Section("DEFAULT").Key("example_key").SetValue("example_value")
		err = m.data.SaveTo(m.filePath)
		if err != nil {
			log.Fatalf("Failed to create config file: %v", err)
		}
		fmt.Printf("Created new config file: %s\n", m.filePath)
	} else {
		// 如果文件存在，加载配置文件
		var err error
		m.data, err = ini.Load(m.filePath)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}
	}
}

// saveConfig saves the INI configuration to the file.
func (m *ConfigManager) saveConfig() {
	m.mu.Lock()
	m.isManualUpdate = true
	defer func() {
		m.isManualUpdate = false
		m.mu.Unlock()
	}()

	err := m.data.SaveTo(m.filePath)
	if err != nil {
		log.Printf("Failed to save config to file: %v", err)
	}
}

// ReadConfig reads a value from the configuration.
func ReadConfig(section, key string) string {
	if instance == nil {
		return ""
	}
	instance.mu.RLock()
	defer instance.mu.RUnlock()
	if instance.data == nil {
		return ""
	}

	// 检查节是否存在
	if !instance.data.HasSection(section) {
		return "" // 如果节不存在，返回空字符串
	}

	// 检查键是否存在于该节中
	s := instance.data.Section(section)
	if !s.HasKey(key) {
		return "" // 如果键不存在，返回空字符串
	}

	// 安全返回键的值
	return s.Key(key).String()
}

// WriteConfig writes a value to the configuration.
func WriteConfig(section, key, value string) {
	cm := GetInstance()
	cm.mu.Lock()
	if cm.data == nil {
		cm.data = ini.Empty()
	}
	cm.data.Section(section).Key(key).SetValue(value)
	cm.mu.Unlock()
	cm.saveConfig()
}

// watchConfig starts a goroutine to watch the configuration file for changes.
// func (m *ConfigManager) watchConfig() {
// 	var err error
// 	m.watcher, err = fsnotify.NewWatcher()
// 	if err != nil {
// 		log.Fatalf("Failed to create watcher: %v", err)
// 	}

// 	go func() {
// 		defer m.watcher.Close()
// 		for {
// 			select {
// 			case event, ok := <-m.watcher.Events:
// 				if !ok {
// 					return
// 				}
// 				if event.Op&fsnotify.Write == fsnotify.Write {
// 					if !m.isManualUpdate {
// 						log.Println("Detected config file change. Reloading...")
// 						m.reloadConfigWithDelay()
// 					}
// 				}
// 			case err, ok := <-m.watcher.Errors:
// 				if !ok {
// 					return
// 				}
// 				log.Printf("Watcher error: %v", err)
// 			}
// 		}
// 	}()

// 	err = m.watcher.Add(m.filePath)
// 	if err != nil {
// 		log.Fatalf("Failed to add watcher to file: %v", err)
// 	}
// }

// reloadConfigWithDelay reloads the configuration file with a delay to avoid rapid successive reloads.
// func (m *ConfigManager) reloadConfigWithDelay() {
// 	time.Sleep(100 * time.Millisecond) // 等待100毫秒以避免频繁重新加载
// 	m.loadConfig()
// }
