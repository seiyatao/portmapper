package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/svc"

	"portmapper/internal/config"
	"portmapper/internal/logging"
	"portmapper/internal/manager"
	"portmapper/internal/service"
	"portmapper/internal/util"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		// 判断是否在交互式会话中运行 (即是否作为 Windows 服务运行)
		isInteractive, err := svc.IsAnInteractiveSession()
		if err != nil {
			fmt.Printf("无法判断是否在交互式会话中运行: %v\n", err)
			os.Exit(1)
		}
		if !isInteractive {
			// 作为 Windows 服务运行
			exePath, _ := os.Executable()
			defaultCfgPath := filepath.Join(filepath.Dir(exePath), "config.json")
			if err := service.RunService(defaultCfgPath); err != nil {
				logging.Error("服务运行失败: %v", err)
			}
			return
		}
		printUsage()
		return
	}

	// 处理命令行指令
	cmd := args[0]
	switch cmd {
	case "install":
		err := service.InstallService("PortMapper", "内部 TCP/UDP 端口映射服务")
		if err != nil {
			fmt.Printf("安装服务失败: %v\n", err)
		} else {
			fmt.Println("服务安装成功。")
		}
	case "uninstall":
		err := service.UninstallService("PortMapper")
		if err != nil {
			fmt.Printf("卸载服务失败: %v\n", err)
		} else {
			fmt.Println("服务卸载成功。")
		}
	case "start":
		err := service.StartService("PortMapper")
		if err != nil {
			fmt.Printf("启动服务失败: %v\n", err)
		} else {
			fmt.Println("服务启动成功。")
		}
	case "stop":
		err := service.StopService("PortMapper")
		if err != nil {
			fmt.Printf("停止服务失败: %v\n", err)
		} else {
			fmt.Println("服务停止成功。")
		}
	case "run":
		// 以前台模式运行，方便调试
		runForeground(*configPath)
	default:
		printUsage()
	}
}

// runForeground 以前台控制台模式运行服务
func runForeground(cfgPath string) {
	// 1. 加载配置
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 校验配置
	if err := util.ValidateConfig(cfg); err != nil {
		fmt.Printf("配置校验失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 初始化日志 (前台模式输出到控制台)
	logging.InitLogger("", false)
	logging.Info("以前台模式启动")

	// 4. 启动所有映射规则
	mgr := manager.NewManager(cfg)
	mgr.StartAll()

	// 5. 监听退出信号，实现优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logging.Info("正在关闭...")
	mgr.StopAll()
	logging.Info("关闭完成")
}

func printUsage() {
	fmt.Println("用法: portmapper [选项] <命令>")
	fmt.Println("命令:")
	fmt.Println("  install   - 安装为 Windows 服务")
	fmt.Println("  uninstall - 卸载 Windows 服务")
	fmt.Println("  start     - 启动 Windows 服务")
	fmt.Println("  stop      - 停止 Windows 服务")
	fmt.Println("  run       - 以前台模式运行 (控制台模式)")
	fmt.Println("选项:")
	fmt.Println("  -config   - 配置文件路径 (默认: config.json)")
}
