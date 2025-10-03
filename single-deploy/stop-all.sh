#!/bin/bash

# Mini-HIDS 停止脚本

echo "=== Mini-HIDS 停止脚本 ==="

# 查找并停止 mini-hids-server-linux 进程
echo "🔍 查找 Mini-HIDS Server 进程..."
SERVER_PIDS=$(pgrep -f "mini-hids-server-linux")

if [ -n "$SERVER_PIDS" ]; then
    echo "🛑 停止 Server 进程: $SERVER_PIDS"
    kill $SERVER_PIDS
    sleep 2
    
    # 检查是否还有残留进程
    REMAINING_SERVER=$(pgrep -f "mini-hids-server-linux")
    if [ -n "$REMAINING_SERVER" ]; then
        echo "⚠️  强制停止残留 Server 进程: $REMAINING_SERVER"
        kill -9 $REMAINING_SERVER
    fi
    echo "✅ Server 已停止"
else
    echo "ℹ️  未找到运行中的 Server 进程"
fi

# 查找并停止 mini-hids-agent-linux 进程
echo "🔍 查找 Mini-HIDS Agent 进程..."
AGENT_PIDS=$(pgrep -f "mini-hids-agent-linux")

if [ -n "$AGENT_PIDS" ]; then
    echo "🛑 停止 Agent 进程: $AGENT_PIDS"
    kill $AGENT_PIDS
    sleep 2
    
    # 检查是否还有残留进程
    REMAINING_AGENT=$(pgrep -f "mini-hids-agent-linux")
    if [ -n "$REMAINING_AGENT" ]; then
        echo "⚠️  强制停止残留 Agent 进程: $REMAINING_AGENT"
        kill -9 $REMAINING_AGENT
    fi
    echo "✅ Agent 已停止"
else
    echo "ℹ️  未找到运行中的 Agent 进程"
fi

echo ""
echo "🎉 Mini-HIDS 已完全停止!"