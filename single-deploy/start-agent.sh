#!/bin/bash

# Mini-HIDS Agent 启动脚本

echo "=== Mini-HIDS Agent 启动脚本 ==="

# 检查配置文件
if [ ! -f "agent-config.json" ]; then
    echo "❌ 错误: 找不到 agent-config.json 配置文件"
    exit 1
fi

# 检查可执行文件
if [ ! -f "mini-hids-agent-linux" ]; then
    echo "❌ 错误: 找不到 mini-hids-agent-linux 可执行文件"
    exit 1
fi

# 设置执行权限
chmod +x mini-hids-agent-linux

echo "✅ 配置检查完成"
echo "📊 Agent配置:"
echo "   - 服务器地址: 127.0.0.1:8848"
echo "   - 上报间隔: 30秒"
echo "   - 日志级别: info"
echo "   - 监控进程: 是"
echo "   - 监控文件: 是"
echo "   - 监控网络: 是"
echo "   - 监控系统: 是"
echo ""

echo "🚀 启动 Mini-HIDS Agent..."
./mini-hids-agent-linux -config agent-config.json