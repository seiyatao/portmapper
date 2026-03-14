package config

import (
	"encoding/json"
	"os"
)

// Rule 定义了单条端口映射规则
type Rule struct {
	Name           string `json:"name"`            // 规则名称
	Enabled        bool   `json:"enabled"`         // 是否启用
	Protocol       string `json:"protocol"`        // 协议类型: tcp 或 udp
	Listen         string `json:"listen"`          // 本地监听地址 (例如: 0.0.0.0:8080)
	Target         string `json:"target"`          // 目标转发地址 (例如: 192.168.1.100:80)
	TimeoutSeconds int    `json:"timeout_seconds"` // 超时时间(秒)
	MaxConnections int    `json:"max_connections"` // 最大并发连接数/会话数
}

// Config 定义了整个服务的配置结构
type Config struct {
	ServiceName string `json:"service_name"` // Windows 服务名称
	LogPath     string `json:"log_path"`     // 日志文件路径
	Rules       []Rule `json:"rules"`        // 映射规则列表
}

// LoadConfig 从指定路径读取并解析 JSON 配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
