package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// AgentData 代理数据结构
type AgentData struct {
	AgentID   string                 `json:"agent_id"`
	Hostname  string                 `json:"hostname"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Server 服务器结构
type Server struct {
	port      int
	mux       *http.ServeMux
	dataStore map[string][]AgentData // 简单的内存存储
}

// NewServer 创建新的服务器
func NewServer(port int) *Server {
	mux := http.NewServeMux()
	
	server := &Server{
		port:      port,
		mux:       mux,
		dataStore: make(map[string][]AgentData),
	}
	
	server.setupRoutes()
	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// API 路由
	s.mux.HandleFunc("/api/agent/data", s.corsMiddleware(s.handleAgentData))
	s.mux.HandleFunc("/api/agents", s.corsMiddleware(s.handleGetAgents))
	s.mux.HandleFunc("/api/agents/", s.corsMiddleware(s.handleGetAgentData))
	s.mux.HandleFunc("/api/stats", s.corsMiddleware(s.handleGetStats))
	s.mux.HandleFunc("/api/health", s.corsMiddleware(s.handleHealth))
	
	// 静态文件服务
	s.mux.HandleFunc("/", s.handleIndex)
}

// corsMiddleware CORS 中间件
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next(w, r)
	}
}

// handleAgentData 处理代理数据
func (s *Server) handleAgentData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var agentData AgentData
	
	if err := json.NewDecoder(r.Body).Decode(&agentData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	
	// 设置时间戳
	agentData.Timestamp = time.Now()
	
	// 存储数据
	s.storeAgentData(agentData)
	
	log.Printf("Received data from agent %s (%s)", agentData.AgentID, agentData.Hostname)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// handleGetAgents 获取代理列表
func (s *Server) handleGetAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	agents := make([]map[string]interface{}, 0)
	
	for agentID, dataList := range s.dataStore {
		if len(dataList) > 0 {
			lastData := dataList[len(dataList)-1]
			agent := map[string]interface{}{
				"agent_id":   agentID,
				"hostname":   lastData.Hostname,
				"last_seen":  lastData.Timestamp,
				"data_count": len(dataList),
				"status":     s.getAgentStatus(lastData.Timestamp),
			}
			agents = append(agents, agent)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
}

// handleGetAgentData 获取特定代理的数据
func (s *Server) handleGetAgentData(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 从 URL 路径中提取 agent ID
	path := strings.TrimPrefix(r.URL.Path, "/api/agents/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "data" {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	
	agentID := parts[0]
	
	dataList, exists := s.dataStore[agentID]
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}
	
	// 获取最近的数据（最多 100 条）
	start := 0
	if len(dataList) > 100 {
		start = len(dataList) - 100
	}
	
	recentData := dataList[start:]
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"data":     recentData,
		"total":    len(dataList),
	})
}

// handleGetStats 获取系统统计信息
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	stats := map[string]interface{}{
		"total_agents":  len(s.dataStore),
		"active_agents": s.getActiveAgentCount(),
		"total_records": s.getTotalRecordCount(),
		"server_uptime": time.Since(startTime).String(),
		"last_updated":  time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleHealth 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// handleIndex 处理首页
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mini-HIDS Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .stat-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stat-value { font-size: 2em; font-weight: bold; color: #2196F3; }
        .stat-label { color: #666; margin-top: 5px; }
        .agents { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .agent-item { padding: 10px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
        .status { padding: 4px 8px; border-radius: 4px; color: white; font-size: 0.8em; }
        .status.online { background-color: #4CAF50; }
        .status.warning { background-color: #FF9800; }
        .status.offline { background-color: #F44336; }
        .refresh-btn { background: #2196F3; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
        .refresh-btn:hover { background: #1976D2; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Mini-HIDS Dashboard</h1>
            <p>主机入侵检测系统 - 实时监控面板</p>
            <button class="refresh-btn" onclick="loadData()">刷新数据</button>
        </div>
        
        <div class="stats" id="stats">
            <!-- 统计信息将在这里显示 -->
        </div>
        
        <div class="agents">
            <h2>代理列表</h2>
            <div id="agents-list">
                <!-- 代理列表将在这里显示 -->
            </div>
        </div>
    </div>

    <script>
        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const stats = await response.json();
                
                document.getElementById('stats').innerHTML = ` + "`" + `
                    <div class="stat-card">
                        <div class="stat-value">${stats.total_agents}</div>
                        <div class="stat-label">总代理数</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${stats.active_agents}</div>
                        <div class="stat-label">活跃代理</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${stats.total_records}</div>
                        <div class="stat-label">总记录数</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">${stats.server_uptime}</div>
                        <div class="stat-label">运行时间</div>
                    </div>
                ` + "`" + `;
            } catch (error) {
                console.error('Failed to load stats:', error);
            }
        }
        
        async function loadAgents() {
            try {
                const response = await fetch('/api/agents');
                const data = await response.json();
                
                const agentsList = document.getElementById('agents-list');
                if (data.agents && data.agents.length > 0) {
                    agentsList.innerHTML = data.agents.map(agent => ` + "`" + `
                        <div class="agent-item">
                            <div>
                                <strong>${agent.hostname}</strong> (${agent.agent_id})
                                <br>
                                <small>最后上报: ${new Date(agent.last_seen).toLocaleString()}</small>
                            </div>
                            <div>
                                <span class="status ${agent.status}">${agent.status}</span>
                                <small>${agent.data_count} 条记录</small>
                            </div>
                        </div>
                    ` + "`" + `).join('');
                } else {
                    agentsList.innerHTML = '<p>暂无代理连接</p>';
                }
            } catch (error) {
                console.error('Failed to load agents:', error);
                document.getElementById('agents-list').innerHTML = '<p>加载失败</p>';
            }
        }
        
        function loadData() {
            loadStats();
            loadAgents();
        }
        
        // 页面加载时获取数据
        loadData();
        
        // 每30秒自动刷新
        setInterval(loadData, 30000);
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// storeAgentData 存储代理数据
func (s *Server) storeAgentData(data AgentData) {
	// 简单的内存存储，实际项目中应该使用数据库
	if s.dataStore[data.AgentID] == nil {
		s.dataStore[data.AgentID] = make([]AgentData, 0)
	}
	
	s.dataStore[data.AgentID] = append(s.dataStore[data.AgentID], data)
	
	// 保持最近的 1000 条记录
	if len(s.dataStore[data.AgentID]) > 1000 {
		s.dataStore[data.AgentID] = s.dataStore[data.AgentID][len(s.dataStore[data.AgentID])-1000:]
	}
}

// getAgentStatus 获取代理状态
func (s *Server) getAgentStatus(lastSeen time.Time) string {
	timeSince := time.Since(lastSeen)
	
	if timeSince < 30*time.Second {
		return "online"
	} else if timeSince < 5*time.Minute {
		return "warning"
	} else {
		return "offline"
	}
}

// getActiveAgentCount 获取活跃代理数量
func (s *Server) getActiveAgentCount() int {
	count := 0
	threshold := time.Now().Add(-5 * time.Minute)
	
	for _, dataList := range s.dataStore {
		if len(dataList) > 0 {
			lastData := dataList[len(dataList)-1]
			if lastData.Timestamp.After(threshold) {
				count++
			}
		}
	}
	
	return count
}

// getTotalRecordCount 获取总记录数
func (s *Server) getTotalRecordCount() int {
	total := 0
	for _, dataList := range s.dataStore {
		total += len(dataList)
	}
	return total
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

var startTime time.Time

func main() {
	startTime = time.Now()
	
	// 创建服务器
	server := NewServer(8080)
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 在 goroutine 中启动服务器
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	log.Println("Mini-HIDS Server started successfully")
	log.Println("Dashboard: http://localhost:8080")
	log.Println("API endpoints:")
	log.Println("  POST /api/agent/data     - Receive agent data")
	log.Println("  GET  /api/agents         - Get agent list")
	log.Println("  GET  /api/agents/:id/data - Get agent data")
	log.Println("  GET  /api/stats          - Get system stats")
	log.Println("  GET  /api/health         - Health check")
	
	// 等待信号
	<-sigChan
	log.Println("Shutting down server...")
}