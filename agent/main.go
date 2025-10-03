package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mini-hids/agent/collector"
	"mini-hids/agent/config"
)

// Agent 主结构体
type Agent struct {
	config    *config.Config
	collector *collector.Collector
	stopCh    chan struct{}
}

// AgentData 上报数据结构
type AgentData struct {
	AgentID   string                 `json:"agent_id"`
	Hostname  string                 `json:"hostname"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewAgent 创建新的 Agent 实例
func NewAgent() *Agent {
	cfg := config.Load()
	return &Agent{
		config:    cfg,
		collector: collector.New(cfg),
		stopCh:    make(chan struct{}),
	}
}

// Start 启动 Agent
func (a *Agent) Start() error {
	log.Println("Starting Mini-HIDS Agent...")

	// 启动数据采集器
	go a.collector.Start(a.stopCh)

	// 启动数据上报协程
	go a.startReporting()

	// 等待停止信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Received stop signal, shutting down...")
	close(a.stopCh)

	return nil
}

// startReporting 启动数据上报
func (a *Agent) startReporting() {
	ticker := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.reportData()
		case <-a.stopCh:
			return
		}
	}
}

// reportData 上报数据到服务端
func (a *Agent) reportData() {
	// 获取采集的数据
	data := a.collector.GetData()
	if len(data) == 0 {
		return
	}

	hostname, _ := os.Hostname()

	// 将所有数据类型合并到一个请求中
	agentData := AgentData{
		AgentID:   hostname, // 使用hostname作为AgentID
		Hostname:  hostname,
		Timestamp: time.Now(),
		Data:      data,
	}

	// 发送到服务端
	if err := a.sendToServer(agentData); err != nil {
		log.Printf("Failed to send data to server: %v", err)
	}
}

// sendToServer 发送数据到服务端
func (a *Agent) sendToServer(data AgentData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%d/api/agent/data", a.config.ServerHost, a.config.ServerPort)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	agent := NewAgent()
	if err := agent.Start(); err != nil {
		log.Fatalf("Agent failed to start: %v", err)
	}
}
