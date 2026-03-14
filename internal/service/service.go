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

// portMapperService 实现 Windows 服务的接口
type portMapperService struct {
	cfgPath string
}

// Execute 是 Windows 服务的核心执行逻辑
func (m *portMapperService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	// 加载配置
	cfg, err := config.LoadConfig(m.cfgPath)
	if err != nil {
		logging.Error("加载配置失败: %v", err)
		return
	}

	// 初始化日志 (服务模式下输出到文件)
	if err := logging.InitLogger(cfg.LogPath, true); err != nil {
		// 日志初始化失败，直接退出服务
		return false, 1
	}
	logging.Info("服务已启动")

	// 启动规则管理器
	mgr := manager.NewManager(cfg)
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
func RunService(name, cfgPath string) error {
	return svc.Run(name, &portMapperService{cfgPath: cfgPath})
}

// InstallService 安装 Windows 服务
func InstallService(name, desc string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("服务 %s 已存在", name)
	}

	s, err = m.CreateService(name, exePath, mgr.Config{
		DisplayName: name,
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
func UninstallService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("服务 %s 未安装", name)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	return nil
}

// StartService 启动已安装的 Windows 服务
func StartService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
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
func StopService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
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
