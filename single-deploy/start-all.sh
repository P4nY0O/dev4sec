#!/bin/bash

# Mini-HIDS 一键启动脚本 (Server + Agent)

echo "=== Mini-HIDS 一键启动脚本 ==="

# 检查所有必要文件
echo "🔍 检查文件..."
missing_files=()

if [ ! -f "mini-hids-server-linux" ]; then
    missing_files+=("mini-hids-server-linux")
fi

if [ ! -f "mini-hids-agent-linux" ]; then
    missing_files+=("mini-hids-agent-linux")
fi

if [ ! -f "server-config.json" ]; then
    missing_files+=("server-config.json")
fi

if [ ! -f "agent-config.json" ]; then
    missing_files+=("agent-config.json")
fi

if [ ! -d "web" ]; then
    missing_files+=("web/")
fi

if [ ${#missing_files[@]} -ne 0 ]; then
    echo "❌ 错误: 缺少以下文件:"
    for file in "${missing_files[@]}"; do
        echo "   - $file"
    done
    exit 1
fi

# 设置执行权限
chmod +x mini-hids-server-linux mini-hids-agent-linux

echo "✅ 文件检查完成"
echo ""

# 创建日志目录
mkdir -p logs

echo "🚀 启动 Mini-HIDS Server (后台运行)..."
nohup ./mini-hids-server-linux -config server-config.json > logs/server.log 2>&1 &
SERVER_PID=$!
echo "   Server PID: $SERVER_PID"

# 等待服务器启动
echo "⏳ 等待服务器启动..."
sleep 3

# 检查服务器是否启动成功
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✅ Server 启动成功"
else
    echo "❌ Server 启动失败，请检查 logs/server.log"
    exit 1
fi

echo ""
echo "🚀 启动 Mini-HIDS Agent (后台运行)..."
nohup ./mini-hids-agent-linux -config agent-config.json > logs/agent.log 2>&1 &
AGENT_PID=$!
echo "   Agent PID: $AGENT_PID"

# 等待Agent启动
sleep 2

# 检查Agent是否启动成功
if kill -0 $AGENT_PID 2>/dev/null; then
    echo "✅ Agent 启动成功"
else
    echo "❌ Agent 启动失败，请检查 logs/agent.log"
    exit 1
fi

echo ""
echo "🎉 Mini-HIDS 启动完成!"
echo ""
echo "📊 服务信息:"
echo "   - Web界面: http://localhost:8848"
echo "   - Server PID: $SERVER_PID"
echo "   - Agent PID: $AGENT_PID"
echo ""
echo "📝 日志文件:"
echo "   - Server日志: logs/server.log"
echo "   - Agent日志: logs/agent.log"
echo ""
echo "🛑 停止服务:"
echo "   kill $SERVER_PID $AGENT_PID"
echo ""
echo "💡 提示: 使用 'tail -f logs/server.log' 查看服务器日志"
echo "💡 提示: 使用 'tail -f logs/agent.log' 查看Agent日志"