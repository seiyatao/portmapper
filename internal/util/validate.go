package util

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"portmapper/internal/config"
)

// ValidateConfig 校验配置文件的合法性
// 包括必填项检查、协议检查、地址格式检查以及端口冲突检查
func ValidateConfig(cfg *config.Config) error {
	if cfg.ServiceName == "" {
		return errors.New("service_name 是必填项")
	}

	// 用于检测是否存在重复的监听地址
	listenMap := make(map[string]string)

	for _, rule := range cfg.Rules {
		if !rule.Enabled {
			continue
		}

		if rule.Name == "" {
			return errors.New("规则名称(name)是必填项")
		}

		protocol := strings.ToLower(rule.Protocol)
		if protocol != "tcp" && protocol != "udp" {
			return fmt.Errorf("规则 %s: 协议必须是 tcp 或 udp", rule.Name)
		}

		// 校验地址格式
		if protocol == "tcp" {
			if _, err := net.ResolveTCPAddr("tcp", rule.Listen); err != nil {
				return fmt.Errorf("规则 %s: 无效的监听地址 %s", rule.Name, rule.Listen)
			}
			if _, err := net.ResolveTCPAddr("tcp", rule.Target); err != nil {
				return fmt.Errorf("规则 %s: 无效的目标地址 %s", rule.Name, rule.Target)
			}
		} else if protocol == "udp" {
			if _, err := net.ResolveUDPAddr("udp", rule.Listen); err != nil {
				return fmt.Errorf("规则 %s: 无效的监听地址 %s", rule.Name, rule.Listen)
			}
			if _, err := net.ResolveUDPAddr("udp", rule.Target); err != nil {
				return fmt.Errorf("规则 %s: 无效的目标地址 %s", rule.Name, rule.Target)
			}
		}

		// 检查监听端口冲突
		key := fmt.Sprintf("%s://%s", protocol, rule.Listen)
		if existingRule, exists := listenMap[key]; exists {
			return fmt.Errorf("规则 %s: 监听地址 %s 与规则 %s 冲突", rule.Name, rule.Listen, existingRule)
		}
		listenMap[key] = rule.Name
	}

	return nil
}
