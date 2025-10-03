package collector

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"mini-hids/agent/config"
)

// Collector 数据采集器
type Collector struct {
	config  *config.Config         // 配置信息
	data    map[string]interface{} // 采集到的数据
	dataMux sync.RWMutex           // 数据读写锁
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID     int    `json:"pid"`     // 进程ID
	Name    string `json:"name"`    // 进程名称
	Cmdline string `json:"cmdline"` // 进程命令行
	User    string `json:"user"`    // 进程所属用户
	CPU     string `json:"cpu"`     // CPU占用率
	Memory  string `json:"memory"`  // 内存占用率
}

// NetworkConnection 网络连接信息
type NetworkConnection struct {
	Protocol   string `json:"protocol"`    // 协议类型（TCP/UDP）
	LocalAddr  string `json:"local_addr"`  // 本地地址
	LocalPort  int    `json:"local_port"`  // 本地端口
	RemoteAddr string `json:"remote_addr"` // 远程地址
	RemotePort int    `json:"remote_port"` // 远程端口
	State      string `json:"state"`       // 连接状态
	PID        int    `json:"pid"`         // 进程ID
}

// SystemInfo 系统信息
type SystemInfo struct {
	Hostname    string  `json:"hostname"`     // 主机名
	OS          string  `json:"os"`           // 操作系统
	Kernel      string  `json:"kernel"`       // 内核版本
	Uptime      string  `json:"uptime"`       // 系统运行时间
	LoadAverage string  `json:"load_average"` // 系统负载平均值
	CPUUsage    float64 `json:"cpu_usage"`    // CPU占用率
	MemoryUsage float64 `json:"memory_usage"` // 内存占用率
	DiskUsage   float64 `json:"disk_usage"`   // 磁盘占用率
}

// New 创建新的采集器
func New(cfg *config.Config) *Collector {
	return &Collector{
		config: cfg,
		data:   make(map[string]interface{}),
	}
}

// Start 启动采集器
func (c *Collector) Start(stopCh <-chan struct{}) {
	log.Println("Starting data collector...")

	// 定时采集数据
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.collectData()
		case <-stopCh:
			return
		}
	}
}

// collectData 采集各种数据
func (c *Collector) collectData() {
	c.dataMux.Lock()
	defer c.dataMux.Unlock()

	if c.config.CollectProcess {
		c.data["processes"] = c.collectProcesses()
	}

	if c.config.CollectNetwork {
		c.data["network"] = c.collectNetworkConnections()
	}

	if c.config.CollectSystem {
		c.data["system"] = c.collectSystemInfo()
	}
}

// collectProcesses 采集进程信息
func (c *Collector) collectProcesses() []ProcessInfo {
	var processes []ProcessInfo

	procDir := "/proc"
	files, err := os.ReadDir(procDir)
	if err != nil {
		log.Printf("Failed to read /proc: %v", err)
		return processes
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(file.Name())
		if err != nil {
			continue
		}

		process := c.getProcessInfo(pid)
		if process != nil {
			processes = append(processes, *process)
		}
	}

	return processes
}

// getProcessInfo 获取单个进程信息
func (c *Collector) getProcessInfo(pid int) *ProcessInfo {
	// 读取进程名称
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := os.ReadFile(commPath)
	if err != nil {
		return nil
	}
	name := strings.TrimSpace(string(commData))

	// 读取命令行
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineData, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return nil
	}
	cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")

	return &ProcessInfo{
		PID:     pid,
		Name:    name,
		Cmdline: strings.TrimSpace(cmdline),
		User:    "unknown", // 简化版本，不获取用户信息
		CPU:     "0%",      // 简化版本，不计算 CPU 使用率
		Memory:  "0MB",     // 简化版本，不计算内存使用
	}
}

// collectNetworkConnections 采集网络连接信息
func (c *Collector) collectNetworkConnections() []NetworkConnection {
	var connections []NetworkConnection

	// 读取 TCP 连接
	tcpConnections := c.parseNetworkFile("/proc/net/tcp")
	connections = append(connections, tcpConnections...)

	// 读取 UDP 连接
	udpConnections := c.parseNetworkFile("/proc/net/udp")
	connections = append(connections, udpConnections...)

	return connections
}

// parseNetworkFile 解析网络连接文件
func (c *Collector) parseNetworkFile(filePath string) []NetworkConnection {
	var connections []NetworkConnection

	file, err := os.Open(filePath)
	if err != nil {
		return connections
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 跳过标题行
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// 解析本地地址和端口
		localAddr, localPort := c.parseAddress(fields[1])
		// 解析远程地址和端口
		remoteAddr, remotePort := c.parseAddress(fields[2])

		protocol := "tcp"
		if strings.Contains(filePath, "udp") {
			protocol = "udp"
		}

		connection := NetworkConnection{
			Protocol:   protocol,
			LocalAddr:  localAddr,
			LocalPort:  localPort,
			RemoteAddr: remoteAddr,
			RemotePort: remotePort,
			State:      c.getConnectionState(fields[3]),
			PID:        0, // 简化版本，不获取 PID
		}

		connections = append(connections, connection)
	}

	return connections
}

// parseAddress 解析地址和端口
func (c *Collector) parseAddress(addr string) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "0.0.0.0", 0
	}

	// 解析 IP 地址（十六进制格式）
	ipHex := parts[0]
	ip := c.hexToIP(ipHex)

	// 解析端口（十六进制格式）
	portHex := parts[1]
	port, _ := strconv.ParseInt(portHex, 16, 32)

	return ip, int(port)
}

// hexToIP 将十六进制字符串转换为 IP 地址
func (c *Collector) hexToIP(hexStr string) string {
	if len(hexStr) != 8 {
		return "0.0.0.0"
	}

	ip := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		byteHex := hexStr[i*2 : i*2+2]
		byteVal, _ := strconv.ParseUint(byteHex, 16, 8)
		ip[3-i] = byte(byteVal) // 小端序
	}

	return ip.String()
}

// getConnectionState 获取连接状态
func (c *Collector) getConnectionState(stateHex string) string {
	state, _ := strconv.ParseInt(stateHex, 16, 32)
	states := map[int64]string{
		1:  "ESTABLISHED",
		2:  "SYN_SENT",
		3:  "SYN_RECV",
		4:  "FIN_WAIT1",
		5:  "FIN_WAIT2",
		6:  "TIME_WAIT",
		7:  "CLOSE",
		8:  "CLOSE_WAIT",
		9:  "LAST_ACK",
		10: "LISTEN",
		11: "CLOSING",
	}

	if stateName, exists := states[state]; exists {
		return stateName
	}
	return "UNKNOWN"
}

// collectSystemInfo 采集系统信息
func (c *Collector) collectSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()

	return SystemInfo{
		Hostname:    hostname,
		OS:          c.getOSInfo(),
		Kernel:      c.getKernelVersion(),
		Uptime:      c.getUptime(),
		LoadAverage: c.getLoadAverage(),
		CPUUsage:    0.0, // 简化版本
		MemoryUsage: 0.0, // 简化版本
		DiskUsage:   0.0, // 简化版本
	}
}

// getOSInfo 获取操作系统信息
func (c *Collector) getOSInfo() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Unknown"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
		}
	}

	return "Linux"
}

// getKernelVersion 获取内核版本
func (c *Collector) getKernelVersion() string {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fields[2]
	}

	return "Unknown"
}

// getUptime 获取系统运行时间
func (c *Collector) getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 1 {
		uptime, _ := strconv.ParseFloat(fields[0], 64)
		duration := time.Duration(uptime) * time.Second
		return duration.String()
	}

	return "Unknown"
}

// getLoadAverage 获取系统负载
func (c *Collector) getLoadAverage() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "Unknown"
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fmt.Sprintf("%s %s %s", fields[0], fields[1], fields[2])
	}

	return "Unknown"
}

// GetData 获取采集的数据
func (c *Collector) GetData() map[string]interface{} {
	c.dataMux.RLock()
	defer c.dataMux.RUnlock()

	// 复制数据
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}

	return result
}
