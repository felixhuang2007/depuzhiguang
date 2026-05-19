#!/bin/bash
set -e

echo "=== 德扑之光 - 腾讯云部署脚本 ==="

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "正在安装 Docker..."
    sudo apt update
    sudo apt install -y docker.io docker-compose
    sudo systemctl enable docker
    sudo systemctl start docker
    sudo usermod -aG docker $USER
    echo "Docker 安装完成，请重新登录或执行 'newgrp docker'"
fi

# 创建环境变量文件（如果不存在）
if [ ! -f ../apps/api-server/.env ]; then
    echo "创建 api-server .env 文件..."
    cat > ../apps/api-server/.env <<'EOF'
DATABASE_URL=postgresql://depg:depg_pass@postgres:5432/depg_db
JWT_SECRET=change-me-in-production-32-chars-min
JWT_REFRESH_SECRET=change-me-too-in-production-32-chars
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
NODE_ENV=production
PORT=3000
API_BASE_URL=http://localhost:3000
KBZPAY_MERCHANT_ID=test_merchant
KBZPAY_SECRET=test_secret
EOF
fi

# 拉取基础镜像（加速构建）
echo "拉取基础镜像..."
docker pull golang:1.23-alpine
docker pull node:20-alpine
docker pull postgres:16-alpine
docker pull redis:7-alpine

# 构建并启动
echo "构建并启动服务..."
cd "$(dirname "$0")"
docker-compose down 2>/dev/null || true
docker-compose up -d --build

echo ""
echo "=== 部署完成 ==="
echo ""
echo "查看服务状态: docker-compose ps"
echo "查看 game-server 日志: docker-compose logs -f game-server"
echo "查看 bot-service 日志: docker-compose logs -f bot-service"
echo ""
echo "API Server: http://$(curl -s ifconfig.me):3000"
echo "Game Server WS: ws://$(curl -s ifconfig.me):8080/ws"
