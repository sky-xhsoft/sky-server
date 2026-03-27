#!/bin/bash
# Sky-Server 服务器初始化脚本 - 全新空白服务器
# 执行此脚本会自动安装所有依赖: MySQL 8.0, Redis, Go, sky-server

set -e

echo "=========================================="
echo "   Sky-Server 生产环境初始化"
echo "=========================================="
echo ""

# 更新系统
echo ">>> [1/8] 更新系统包..."
apt update -y && apt upgrade -y

# 安装基础工具
echo ">>> [2/8] 安装基础工具..."
apt install -y wget curl git vim net-tools nginx

# 安装 MySQL 8.0
echo ">>> [3/8] 安装 MySQL 8.0..."
apt install -y mysql-server

# 启动 MySQL
systemctl enable mysql
systemctl start mysql

# 安装 Redis
echo ">>> [4/8] 安装 Redis..."
apt install -y redis-server
systemctl enable redis
systemctl start redis

# 安装 Go 1.22
echo ">>> [5/8] 安装 Go 语言环境..."
cd /tmp
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
ln -sf /usr/local/go/bin/go /usr/local/bin/go
ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
go version

# 创建数据库
echo ">>> [6/8] 创建数据库..."
mysql -e "CREATE DATABASE IF NOT EXISTS sky_server DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
echo "数据库 sky_server 创建完成"

# 下载 SQL 初始化脚本
echo ">>> [7/8] 下载并导入初始化 SQL..."
cd /tmp
wget https://raw.githubusercontent.com/sky-xhsoft/sky-server/master/sqls/create_skyserver.sql
mysql sky_server < /tmp/create_skyserver.sql
echo "数据库初始化完成"

# 创建应用目录
echo ">>> [8/8] 部署应用..."
mkdir -p /opt/sky-server/bin
mkdir -p /opt/sky-server/configs
mkdir -p /opt/sky-server/logs

# 复制上传的二进制文件
if [ -f /tmp/sky-server ]; then
    cp /tmp/sky-server /opt/sky-server/bin/
    chmod +x /opt/sky-server/bin/sky-server
fi

# 复制配置文件
if [ -f /tmp/config.yaml ]; then
    cp /tmp/config.yaml /opt/sky-server/configs/
else
    cp /tmp/config.example.yaml /opt/sky-server/configs/config.yaml
fi

# 安装 systemd 服务
if [ -f /tmp/sky-server.service ]; then
    cp /tmp/sky-server.service /etc/systemd/system/
    systemctl daemon-reload
fi

echo ""
echo "=========================================="
echo "       初始化完成!"
echo "=========================================="
echo ""
echo "下一步："
echo "1. 编辑配置文件: nano /opt/sky-server/configs/config.yaml"
echo "   - 修改数据库密码 (如果需要)"
echo "   - 修改腾讯云/阿里云配置 (如果需要直播/云盘功能)"
echo ""
echo "2. 启动服务:"
echo "   systemctl start sky-server"
echo "   systemctl enable sky-server"
echo ""
echo "3. 检查状态:"
echo "   systemctl status sky-server"
echo "   journalctl -u sky-server -f"
echo ""
echo "4. 访问:"
echo "   http://119.45.20.166:9090/swagger/index.html"
echo ""
