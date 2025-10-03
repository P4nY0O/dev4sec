#!/bin/bash

# Mini-HIDS 自动化部署脚本
# 使用方法: ./deploy.sh [服务器IP] [用户名]
# 例如: ./deploy.sh 192.168.1.100 root

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查参数
if [ $# -lt 2 ]; then
    log_error "使用方法: $0 <服务器IP> <用户名>"
    log_error "例如: $0 192.168.1.100 root"
    exit 1
fi

SERVER_IP="$1"
USERNAME="$2"
DEPLOY_PACKAGE="mini-hids-single-deploy.tar.gz"
REMOTE_DIR="/root/mini-hids"

log_info "开始 Mini-HIDS 自动化部署..."
log_info "目标服务器: ${USERNAME}@${SERVER_IP}"
log_info "部署包: ${DEPLOY_PACKAGE}"

# 步骤1: 编译Linux二进制文件
log_info "步骤1: 编译Linux二进制文件..."

log_info "编译Agent..."
cd agent
GOOS=linux GOARCH=amd64 go build -o ../single-deploy/mini-hids-agent-linux
cd ..

log_info "编译Server..."
cd server
GOOS=linux GOARCH=amd64 go build -o ../single-deploy/mini-hids-server-linux
cd ..

log_success "二进制文件编译完成"

# 步骤2: 验证二进制文件
log_info "步骤2: 验证二进制文件..."
if file single-deploy/mini-hids-agent-linux | grep -q "ELF 64-bit LSB executable, x86-64"; then
    log_success "Agent二进制文件格式正确"
else
    log_error "Agent二进制文件格式错误"
    exit 1
fi

if file single-deploy/mini-hids-server-linux | grep -q "ELF 64-bit LSB executable, x86-64"; then
    log_success "Server二进制文件格式正确"
else
    log_error "Server二进制文件格式错误"
    exit 1
fi

# 步骤3: 创建部署包
log_info "步骤3: 创建部署包..."
cd single-deploy
COPYFILE_DISABLE=1 tar -czf ../${DEPLOY_PACKAGE} .
cd ..

if [ -f "${DEPLOY_PACKAGE}" ]; then
    PACKAGE_SIZE=$(ls -lh ${DEPLOY_PACKAGE} | awk '{print $5}')
    log_success "部署包创建成功: ${DEPLOY_PACKAGE} (${PACKAGE_SIZE})"
else
    log_error "部署包创建失败"
    exit 1
fi

# 步骤4: 上传部署包到服务器
log_info "步骤4: 上传部署包到服务器..."
log_info "正在上传 ${DEPLOY_PACKAGE} 到 ${USERNAME}@${SERVER_IP}:${REMOTE_DIR}/"

scp ${DEPLOY_PACKAGE} ${USERNAME}@${SERVER_IP}:${REMOTE_DIR}/
if [ $? -eq 0 ]; then
    log_success "部署包上传成功"
else
    log_error "部署包上传失败"
    exit 1
fi

# 步骤5: 在服务器上执行部署
log_info "步骤5: 在服务器上执行部署..."

ssh ${USERNAME}@${SERVER_IP} << 'EOF'
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

REMOTE_DIR="/root/mini-hids"
DEPLOY_PACKAGE="mini-hids-single-deploy.tar.gz"

cd ${REMOTE_DIR}

log_info "停止现有服务..."
if [ -f "stop-all.sh" ]; then
    chmod +x stop-all.sh
    ./stop-all.sh || log_warning "停止服务时出现警告（可能服务未运行）"
else
    log_warning "stop-all.sh 不存在，跳过停止服务"
fi

log_info "清理旧文件..."
# 备份旧的配置文件
if [ -f "agent-config.json" ]; then
    cp agent-config.json agent-config.json.backup
    log_info "已备份 agent-config.json"
fi

if [ -f "server-config.json" ]; then
    cp server-config.json server-config.json.backup
    log_info "已备份 server-config.json"
fi

# 删除旧文件（保留配置文件备份和日志）
find . -maxdepth 1 -type f ! -name "*.log" ! -name "*.backup" ! -name "${DEPLOY_PACKAGE}" -delete
find . -maxdepth 1 -type d ! -name "." ! -name "logs" -exec rm -rf {} + 2>/dev/null || true

log_info "解压新的部署包..."
tar -xzf ${DEPLOY_PACKAGE}

# 恢复配置文件（如果存在备份）
if [ -f "agent-config.json.backup" ]; then
    if [ ! -f "agent-config.json" ]; then
        cp agent-config.json.backup agent-config.json
        log_info "已恢复 agent-config.json"
    fi
fi

if [ -f "server-config.json.backup" ]; then
    if [ ! -f "server-config.json" ]; then
        cp server-config.json.backup server-config.json
        log_info "已恢复 server-config.json"
    fi
fi

log_info "设置执行权限..."
chmod +x mini-hids-agent-linux mini-hids-server-linux *.sh

log_info "启动新服务..."
./start-all.sh

log_success "部署完成！"

# 等待服务启动
sleep 3

log_info "检查服务状态..."
if pgrep -f "mini-hids-server-linux" > /dev/null; then
    log_success "Server 服务运行正常"
else
    log_error "Server 服务未运行"
fi

if pgrep -f "mini-hids-agent-linux" > /dev/null; then
    log_success "Agent 服务运行正常"
else
    log_error "Agent 服务未运行"
fi

log_info "最近的日志:"
if [ -f "server.log" ]; then
    echo "=== Server 日志 (最后10行) ==="
    tail -10 server.log
fi

if [ -f "agent.log" ]; then
    echo "=== Agent 日志 (最后10行) ==="
    tail -10 agent.log
fi

EOF

if [ $? -eq 0 ]; then
    log_success "服务器部署完成！"
    log_info "你可以通过以下方式访问:"
    log_info "Web界面: http://${SERVER_IP}:8848"
    log_info "API接口: http://${SERVER_IP}:8848/api/stats"
else
    log_error "服务器部署失败"
    exit 1
fi

log_success "🎉 Mini-HIDS 自动化部署完成！"
log_info "部署包已保存在本地: ${DEPLOY_PACKAGE}"
log_info "如需重新部署，直接运行: $0 ${SERVER_IP} ${USERNAME}"