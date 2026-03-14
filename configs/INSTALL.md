# 安装与运行指南

## 1. 准备配置文件

在 `pc-edge-gateway.exe` 同级目录下，创建一个名为 `config.json` 的文件。你可以参考 `configs/config.example.json`：

```json
{
  "service_name": "pc-edge-gateway",
  "service_desc": "内部 TCP/UDP 端口映射服务",
  "log_path": "logs/pc-edge-gateway.log",
  "rules": [
    {
      "name": "web-tcp",
      "enabled": true,
      "protocol": "tcp",
      "listen": "0.0.0.0:8080",
      "target": "192.168.1.100:80",
      "timeout_seconds": 300,
      "max_connections": 1000
    }
  ]
}
```

> **注意**: `service_name` 决定了在 Windows 服务管理器中显示的服务名称。

## 2. 前台运行 (测试)

在正式安装为服务前，建议先以前台模式运行测试配置是否正确：

```cmd
pc-edge-gateway.exe run
```

> **注意**: 程序会自动读取与 `pc-edge-gateway.exe` 同级目录下的 `config.json`。

## 3. 安装为 Windows 服务

使用管理员权限打开命令提示符 (CMD) 或 PowerShell，执行：

```cmd
pc-edge-gateway.exe install
```

安装成功后，启动服务：

```cmd
pc-edge-gateway.exe start
```

## 4. 停止与卸载服务

停止服务：

```cmd
pc-edge-gateway.exe stop
```

卸载服务：

```cmd
pc-edge-gateway.exe uninstall
```
