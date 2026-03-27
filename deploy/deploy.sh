#!/bin/bash
# Sky-Server 生产部署脚本

# 安装目录
INSTALL_DIR="/opt/sky-server"

echo "=== Sky-Server 部署开始 ==="

# 创建目录
mkdir -p $INSTALL_DIR/bin
mkdir -p $INSTALL_DIR/configs
mkdir -p $INSTALL_DIR/logs

# 复制二进制文件
cp bin/sky-server $INSTALL_DIR/bin/
chmod +x $INSTALL_DIR/bin/sky-server

# 复制配置文件
if [ -f configs/config.yaml ]; then
    cp configs/config.yaml $INSTALL_DIR/configs/
else
    cp configs/config.example.yaml $INSTALL_DIR/configs/config.yaml
    echo "⚠️  请编辑 $INSTALL_DIR/configs/config.yaml 修改数据库连接配置！"
fi

# 安装 systemd 服务
cp deploy/sky-server.service /etc/systemd/system/
systemctl daemon-reload

echo ""
echo "=== 部署完成 ==="
echo ""
echo "下一步："
echo "1. 编辑配置文件: nano $INSTALL_DIR/configs/config.yaml"
echo "2. 启动服务: systemctl start sky-server"
echo "3. 设置开机自启: systemctl enable sky-server"
echo "4. 查看状态: systemctl status sky-server"
echo "5. 查看日志: journalctl -u sky-server -f"
echo ""
