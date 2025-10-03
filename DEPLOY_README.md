# Mini-HIDS 自动化部署指南

## 概述

`deploy.sh` 是一个自动化部署脚本，可以一键完成 Mini-HIDS 的编译、打包、上传和部署。

## 功能特性

✅ **自动编译**: 自动编译 Linux 版本的 Agent 和 Server 二进制文件  
✅ **智能打包**: 创建部署包，自动排除 macOS 扩展属性  
✅ **安全上传**: 通过 SCP 安全上传到目标服务器  
✅ **优雅部署**: 停止旧服务、备份配置、部署新版本、启动服务  
✅ **状态检查**: 自动检查服务运行状态和日志  
✅ **配置保护**: 自动备份和恢复配置文件  

## 使用方法

### 基本用法

```bash
./deploy.sh <服务器IP> <用户名>
```

### 示例

```bash
# 部署到 root 用户的服务器
./deploy.sh 192.168.1.100 root

# 部署到普通用户的服务器
./deploy.sh 10.0.0.50 ubuntu
```

## 部署流程

脚本会按以下步骤执行：

### 1. 编译二进制文件
- 编译 Linux 版本的 Agent (`mini-hids-agent-linux`)
- 编译 Linux 版本的 Server (`mini-hids-server-linux`)
- 验证二进制文件格式正确性

### 2. 创建部署包
- 打包 `single-deploy` 目录内容
- 生成 `mini-hids-single-deploy.tar.gz`
- 显示包大小信息

### 3. 上传到服务器
- 通过 SCP 上传部署包到 `/root/mini-hids/`
- 验证上传成功

### 4. 服务器端部署
- 停止现有服务 (`./stop-all.sh`)
- 备份现有配置文件
- 清理旧文件（保留日志和配置备份）
- 解压新部署包
- 恢复配置文件
- 设置执行权限
- 启动新服务 (`./start-all.sh`)

### 5. 状态检查
- 检查 Server 和 Agent 进程状态
- 显示最新日志（最后10行）

## 前置条件

### 本地环境
- Go 开发环境已安装
- 项目代码完整（agent、server、single-deploy 目录）
- SSH 密钥已配置（免密登录目标服务器）

### 目标服务器
- Linux 系统（x86_64 架构）
- SSH 服务已启用
- 目标目录 `/root/mini-hids/` 存在
- 具有执行权限

## 配置文件处理

脚本会智能处理配置文件：

- **部署前**: 自动备份现有的 `agent-config.json` 和 `server-config.json`
- **部署后**: 如果新包中没有配置文件，会自动恢复备份的配置
- **备份文件**: 保存为 `.backup` 后缀（如 `agent-config.json.backup`）

## 日志和故障排除

### 查看部署日志
脚本会实时显示部署过程，包含彩色状态信息：
- 🔵 **[INFO]**: 信息提示
- 🟢 **[SUCCESS]**: 成功操作
- 🟡 **[WARNING]**: 警告信息
- 🔴 **[ERROR]**: 错误信息

### 常见问题

#### 1. SSH 连接失败
```bash
# 确保 SSH 密钥已配置
ssh-copy-id root@192.168.1.100

# 或手动测试连接
ssh root@192.168.1.100
```

#### 2. 编译失败
```bash
# 检查 Go 环境
go version

# 检查项目依赖
go mod tidy
```

#### 3. 服务启动失败
```bash
# 登录服务器检查日志
ssh root@192.168.1.100
cd /root/mini-hids
tail -f server.log
tail -f agent.log
```

#### 4. 权限问题
```bash
# 确保脚本有执行权限
chmod +x deploy.sh

# 确保目标目录权限正确
ssh root@192.168.1.100 "chmod 755 /root/mini-hids"
```

## Web 界面更新

最新版本的 Web 界面已更新，现在会直接在 Dashboard 上显示：

- 📊 **系统信息**: OS、内核、CPU、内存、磁盘使用率
- ⚙️ **主要进程**: 前5个进程的详细信息
- 🌐 **网络连接**: 前5个网络连接状态

无需点击即可查看所有 Agent 的详细信息！

## 访问部署的服务

部署完成后，可以通过以下方式访问：

- **Web 界面**: `http://<服务器IP>:8848`
- **API 接口**: `http://<服务器IP>:8848/api/stats`
- **Agent 列表**: `http://<服务器IP>:8848/api/agents`

## 重新部署

如需重新部署，只需再次运行相同命令：

```bash
./deploy.sh 192.168.1.100 root
```

脚本会自动处理停止旧服务、更新文件、启动新服务的完整流程。

---

**提示**: 首次使用建议先在测试环境验证，确保所有功能正常后再部署到生产环境。