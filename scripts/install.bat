@echo off
chcp 65001 > nul
echo 正在安装 pc-edge-gateway 服务...
pc-edge-gateway.exe install
echo 正在启动 pc-edge-gateway 服务...
pc-edge-gateway.exe start
echo 完成。
pause
