package service

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"pc-edge-gateway/internal/config"
	"pc-edge-gateway/internal/logging"
	"pc-edge-gateway/internal/manager"
)

// edgeGatewayService 实现 Windows 服务的接口
type edgeGatewayService struct {
	cfg *config.Config
}

// Execute 是 Windows 服务的核心执行逻辑
func (m *edgeGatewayService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// 初始化日志 (服务模式下输出到文件)
	if err := logging.InitLogger(m.cfg.LogPath, true); err != nil {
		// 日志初始化失败，直接退出服务
		return false, 1
	}
	logging.Info("服务已启动")

	// 启动规则管理器
	mgr := manager.NewManager(m.cfg)
	startedCount := mgr.StartAll()
	if startedCount == 0 {
		logging.Error("没有成功启动任何规则，服务退出")
		return false, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			logging.Info("服务正在停止")
			mgr.StopAll() // 优雅退出，清理资源
			break loop
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	logging.Info("服务已停止")
	return
}

// RunService 启动 Windows 服务
func RunService(cfg *config.Config) error {
	return svc.Run(cfg.ServiceName, &edgeGatewayService{cfg: cfg})
}

// InstallService 安装 Windows 服务
func InstallService(cfg *config.Config) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.ServiceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("服务 %s 已存在", cfg.ServiceName)
	}

	desc := cfg.ServiceDesc
	if desc == "" {
		desc = "内部 TCP/UDP 端口映射服务"
	}

	s, err = m.CreateService(cfg.ServiceName, exePath, mgr.Config{
		DisplayName: cfg.ServiceName,
		Description: desc,
		StartType:   mgr.StartAutomatic, // 设置为自动启动
	})
	if err != nil {
		return err
	}
	defer s.Close()

	return nil
}

// UninstallService 卸载 Windows 服务
func UninstallService(cfg *config.Config) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("服务 %s 未安装", cfg.ServiceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	return nil
}

// StartService 启动已安装的 Windows 服务
func StartService(cfg *config.Config) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("无法访问服务: %v", err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("无法启动服务: %v", err)
	}

	return nil
}

// StopService 停止正在运行的 Windows 服务
func StopService(cfg *config.Config) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(cfg.ServiceName)
	if err != nil {
		return fmt.Errorf("无法访问服务: %v", err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("无法发送控制指令=%d: %v", svc.Stop, err)
	}

	if status.State != svc.Stopped {
		fmt.Println("服务正在停止...")
	}

	return nil
}
