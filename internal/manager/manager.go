package manager

import (
	"strings"

	"portmapper/internal/config"
	"portmapper/internal/forward"
	"portmapper/internal/logging"
)

// Forwarder 定义了转发器的通用接口
type Forwarder interface {
	Start() error
	Stop()
}

// Manager 负责管理所有映射规则的生命周期
type Manager struct {
	cfg        *config.Config
	forwarders []Forwarder
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg: cfg,
	}
}

// StartAll 启动所有已启用的映射规则，返回成功启动的规则数量
func (m *Manager) StartAll() int {
	count := 0
	for _, rule := range m.cfg.Rules {
		if !rule.Enabled {
			continue
		}

		var f Forwarder
		protocol := strings.ToLower(rule.Protocol)
		if protocol == "tcp" {
			f = forward.NewTCPForwarder(rule)
		} else if protocol == "udp" {
			f = forward.NewUDPForwarder(rule)
		} else {
			continue
		}

		// 单条规则启动失败不影响其他规则
		if err := f.Start(); err != nil {
			logging.Error("规则 %s 启动失败: %v", rule.Name, err)
		} else {
			m.forwarders = append(m.forwarders, f)
			count++
		}
	}
	return count
}

// StopAll 停止所有正在运行的映射规则
func (m *Manager) StopAll() {
	for _, f := range m.forwarders {
		f.Stop()
	}
	m.forwarders = nil
}
