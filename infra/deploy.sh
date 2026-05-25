#!/bin/bash
set -e

echo "=== 德扑之光 - 腾讯云部署脚本 ==="

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    OS=$(uname -s)
fi

echo "检测到操作系统: $OS"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo "正在安装 Docker..."
    if [[ "$OS" == "centos" || "$OS" == "rhel" || "$OS" == "tencentos" || "$OS" == "rocky" || "$OS" == "almalinux" ]]; then
        # RHEL/CentOS/TencentOS 系
        yum install -y yum-utils
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo || true
        yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        systemctl enable docker
        systemctl start docker
    else
        # Debian/Ubuntu 系
        apt-get update
        apt-get install -y docker.io docker-compose-plugin
        systemctl enable docker
        systemctl start docker
    fi
    usermod -aG docker $USER 2>/dev/null || true
    echo "Docker 安装完成"
fi

# 确保 docker compose 可用
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    echo "正在安装 docker-compose..."
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    COMPOSE_CMD="docker-compose"
fi

echo "使用编排命令: $COMPOSE_CMD"

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
$COMPOSE_CMD down 2>/dev/null || true
$COMPOSE_CMD up -d --build

echo ""
echo "=== 部署完成 ==="
echo ""
echo "查看服务状态: $COMPOSE_CMD ps"
echo "查看 game-server 日志: $COMPOSE_CMD logs -f game-server"
echo "查看 bot-service 日志: $COMPOSE_CMD logs -f simulation-service"
echo ""
IP=$(curl -s ifconfig.me 2>/dev/null || echo "<服务器IP>")
echo "API Server: http://$IP:3000"
echo "Game Server WS: ws://$IP:8080/ws"
