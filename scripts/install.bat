@echo off
chcp 65001 > nul
echo 正在安装 PortMapper 服务...
portmapper.exe install
echo 正在启动 PortMapper 服务...
portmapper.exe start
echo 完成。
pause
