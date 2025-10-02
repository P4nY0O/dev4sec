package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config Agent 配置结构
type Config struct {
	ServerHost     string `json:"server_host"`
	ServerPort     int    `json:"server_port"`
	ReportInterval int    `json:"report_interval"` // 上报间隔（秒）
	LogLevel       string `json:"log_level"`

	// 采集配置
	CollectProcess bool `json:"collect_process"`
	CollectFile    bool `json:"collect_file"`
	CollectNetwork bool `json:"collect_network"`
	CollectSystem  bool `json:"collect_system"`

	// 监控路径
	WatchPaths []string `json:"watch_paths"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		ServerHost:     "127.0.0.1",
		ServerPort:     8848,
		ReportInterval: 30,
		LogLevel:       "info",

		CollectProcess: true,
		CollectFile:    true,
		CollectNetwork: true,
		CollectSystem:  true,

		WatchPaths: []string{
			"/etc",
			"/bin",
			"/sbin",
			"/usr/bin",
			"/usr/sbin",
		},
	}
}

// Load 加载配置文件
func Load() *Config {
	configPath := "config.json"

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		config.Save(configPath)
		return config
	}

	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("Failed to read config file: %v, using default config", err)
		return DefaultConfig()
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Failed to parse config file: %v, using default config", err)
		return DefaultConfig()
	}

	return &config
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
