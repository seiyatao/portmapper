# Windows 端口映射服务 (Port Mapper)

一个基于 Go 语言开发的轻量、稳定且易于部署的 Windows TCP/UDP 端口映射服务。

## 功能特点

- 支持 TCP 和 UDP 端口映射
- 基于 JSON 的简单配置
- 支持多条规则并发运行
- 深度集成 Windows 服务 (支持安装、启动、停止、卸载)
- 支持前台调试模式
- 优雅退出与资源回收
- 极低的系统资源占用

## 编译指南

```bash
go build -o pc-edge-gateway.exe ./cmd/pc-edge-gateway
```

## 使用说明

### 配置文件

在可执行文件同级目录下创建 `config.json` 文件（程序固定读取该路径）：

```json
{
  "service_name": "pc-edge-gateway",
  "log_path": "logs/pc-edge-gateway.log",
  "rules": [
    {
      "name": "web-tcp",
      "enabled": true,
      "protocol": "tcp",
      "listen": "0.0.0.0:8080",
      "target": "192.168.1.100:80",
      "timeout_seconds": 300
    },
    {
      "name": "dns-udp",
      "enabled": true,
      "protocol": "udp",
      "listen": "0.0.0.0:5353",
      "target": "192.168.1.101:5353",
      "timeout_seconds": 60
    }
  ]
}
```

> **提示**: 日志文件支持按天自动轮转，例如 `logs/pc-edge-gateway-2023-10-25.log`。相对路径会自动基于可执行文件所在目录进行解析，防止服务模式下路径错误。并且系统会自动清理超过 15 天的旧日志。

### 命令行指令

- 以前台模式运行 (用于测试和调试):
  ```cmd
  pc-edge-gateway.exe run
  ```

- 安装为 Windows 服务:
  ```cmd
  pc-edge-gateway.exe install
  ```

- 启动服务:
  ```cmd
  pc-edge-gateway.exe start
  ```

- 停止服务:
  ```cmd
  pc-edge-gateway.exe stop
  ```

- 卸载服务:
  ```cmd
  pc-edge-gateway.exe uninstall
  ```
