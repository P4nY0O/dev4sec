package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config Agent 配置结构
type Config struct {
	ServerHost     string `json:"server_host"` // 服务器地址
	ServerPort     int    `json:"server_port"` // 服务器端口
	ReportInterval int    `json:"report_interval"` // 上报间隔（秒）
	LogLevel       string `json:"log_level"`       // 日志级别

	// 采集配置
	CollectProcess bool `json:"collect_process"` // 是否采集进程信息
	CollectFile    bool `json:"collect_file"`    // 是否采集文件信息
	CollectNetwork bool `json:"collect_network"` // 是否采集网络信息
	CollectSystem  bool `json:"collect_system"`  // 是否采集系统信息

	// 监控路径
	WatchPaths []string `json:"watch_paths"` // 监控的文件路径列表
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
	data, err := os.ReadFile(configPath)
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

	return os.WriteFile(path, data, 0644)
}
