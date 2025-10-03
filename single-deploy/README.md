# Mini-HIDS 单机部署指南

## 🏗️ 架构说明

本部署包在单台Linux服务器上同时运行Server和Agent：

```
┌─────────────────────────────────┐
│        Linux 服务器              │
│                                 │
│  ┌───────────┐  ┌───────────┐   │
│  │  Server   │  │   Agent   │   │
│  │ (端口8848) │◄─│(本地连接)  │   │
│  └───────────┘  └───────────┘   │
│  ┌───────────┐                  │
│  │Dashboard  │                  │
│  │(Web界面)  │                  │
│  └───────────┘                  │
└─────────────────────────────────┘
```

## 📦 文件说明

- `mini-hids-server-linux` - 服务器可执行文件
- `mini-hids-agent-linux` - Agent可执行文件
- `server-config.json` - 服务器配置文件
- `agent-config.json` - Agent配置文件
- `web/` - Web界面文件目录
- `start-server.sh` - 单独启动服务器
- `start-agent.sh` - 单独启动Agent
- `start-all.sh` - 一键启动所有服务
- `stop-all.sh` - 一键停止所有服务

## 🚀 快速部署

### 1. 上传文件到服务器

```bash
# 方法1: 使用scp上传整个目录
scp -r single-deploy/ root@your-server-ip:/opt/mini-hids/

# 方法2: 先打包再上传
tar -czf mini-hids-single.tar.gz single-deploy/
scp mini-hids-single.tar.gz root@your-server-ip:/opt/
```

### 2. 在服务器上解压和安装

```bash
# 登录服务器
ssh root@your-server-ip

# 如果使用方法2，先解压
cd /opt
tar -xzf mini-hids-single.tar.gz
mv single-deploy mini-hids

# 进入目录
cd /opt/mini-hids

# 设置执行权限
chmod +x *.sh mini-hids-*-linux
```

### 3. 启动服务

```bash
# 一键启动所有服务（推荐）
./start-all.sh

# 或者分别启动
./start-server.sh &  # 后台启动服务器
./start-agent.sh &   # 后台启动Agent
```

### 4. 访问Web界面

打开浏览器访问：`http://your-server-ip:8848`

## ⚙️ 配置说明

### 服务器配置 (server-config.json)

```json
{
  "port": 8848,              // 服务端口
  "log_level": "info",       // 日志级别
  "web_dir": "./web",        // Web文件目录
  "database": {
    "type": "sqlite",        // 数据库类型
    "path": "./mini-hids.db" // 数据库文件路径
  },
  "security": {
    "enable_auth": false,    // 是否启用认证
    "api_key": ""           // API密钥
  }
}
```

### Agent配置 (agent-config.json)

```json
{
  "server_host": "127.0.0.1",    // 服务器地址（本地）
  "server_port": 8848,           // 服务器端口
  "report_interval": 30,         // 上报间隔（秒）
  "log_level": "info",           // 日志级别
  "collect_process": true,       // 收集进程信息
  "collect_file": true,          // 收集文件信息
  "collect_network": true,       // 收集网络信息
  "collect_system": true,        // 收集系统信息
  "watch_paths": [               // 监控路径
    "/etc/passwd",
    "/etc/shadow",
    "/etc/hosts",
    "/var/log/auth.log",
    "/var/log/secure",
    "/home",
    "/root",
    "/tmp"
  ]
}
```

## 🔧 管理命令

### 启动服务

```bash
# 一键启动所有服务
./start-all.sh

# 单独启动服务器
./start-server.sh

# 单独启动Agent
./start-agent.sh
```

### 停止服务

```bash
# 一键停止所有服务
./stop-all.sh

# 手动停止
pkill -f mini-hids-server-linux
pkill -f mini-hids-agent-linux
```

### 查看状态

```bash
# 查看进程状态
ps aux | grep mini-hids

# 查看端口占用
netstat -tlnp | grep 8848

# 查看日志
tail -f logs/server.log
tail -f logs/agent.log
```

## 🔥 防火墙配置

确保服务器防火墙允许8848端口：

```bash
# CentOS/RHEL
firewall-cmd --permanent --add-port=8848/tcp
firewall-cmd --reload

# Ubuntu/Debian
ufw allow 8848/tcp

# 或者直接使用iptables
iptables -A INPUT -p tcp --dport 8848 -j ACCEPT
```

## 🐛 故障排除

### 1. 服务启动失败

```bash
# 检查日志
cat logs/server.log
cat logs/agent.log

# 检查端口占用
netstat -tlnp | grep 8848

# 检查文件权限
ls -la mini-hids-*-linux
```

### 2. Agent连接失败

```bash
# 检查服务器是否启动
ps aux | grep mini-hids-server

# 检查网络连接
telnet 127.0.0.1 8848

# 检查配置文件
cat agent-config.json
```

### 3. Web界面无法访问

```bash
# 检查服务器状态
curl http://localhost:8848

# 检查防火墙
iptables -L | grep 8848

# 检查web目录
ls -la web/
```

## 🔄 开机自启动（可选）

创建systemd服务：

```bash
# 创建服务文件
cat > /etc/systemd/system/mini-hids.service << EOF
[Unit]
Description=Mini-HIDS Security Monitoring System
After=network.target

[Service]
Type=forking
User=root
WorkingDirectory=/opt/mini-hids
ExecStart=/opt/mini-hids/start-all.sh
ExecStop=/opt/mini-hids/stop-all.sh
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 启用服务
systemctl daemon-reload
systemctl enable mini-hids
systemctl start mini-hids
```

## 📊 监控效果

启动成功后，你将看到：

1. **Web界面**: http://your-server-ip:8848
2. **实时监控数据**:
   - 进程监控
   - 文件变化监控
   - 网络连接监控
   - 系统信息监控
3. **日志记录**: logs/目录下的详细日志

## 🆘 技术支持

如遇问题，请检查：
1. 日志文件 (logs/server.log, logs/agent.log)
2. 配置文件格式是否正确
3. 防火墙和网络设置
4. 文件权限设置