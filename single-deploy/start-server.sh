#!/bin/bash

# Mini-HIDS Server 启动脚本

echo "=== Mini-HIDS Server 启动脚本 ==="

# 检查配置文件
if [ ! -f "server-config.json" ]; then
    echo "❌ 错误: 找不到 server-config.json 配置文件"
    exit 1
fi

# 检查可执行文件
if [ ! -f "mini-hids-server-linux" ]; then
    echo "❌ 错误: 找不到 mini-hids-server-linux 可执行文件"
    exit 1
fi

# 检查web目录
if [ ! -d "web" ]; then
    echo "❌ 错误: 找不到 web 目录"
    exit 1
fi

# 设置执行权限
chmod +x mini-hids-server-linux

echo "✅ 配置检查完成"
echo "📊 服务器配置:"
echo "   - 端口: 8848"
echo "   - Web界面: http://localhost:8848"
echo "   - 日志级别: info"
echo ""

echo "🚀 启动 Mini-HIDS Server..."
./mini-hids-server-linux -config server-config.json