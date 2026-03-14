package forward

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"portmapper/internal/config"
	"portmapper/internal/logging"
)

// TCPForwarder 负责处理单条 TCP 映射规则
type TCPForwarder struct {
	rule     config.Rule
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

func NewTCPForwarder(rule config.Rule) *TCPForwarder {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPForwarder{
		rule:   rule,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动 TCP 监听
func (f *TCPForwarder) Start() error {
	listener, err := net.Listen("tcp", f.rule.Listen)
	if err != nil {
		return err
	}
	f.listener = listener

	logging.Info("规则 %s 正在监听 %s -> %s (TCP)", f.rule.Name, f.rule.Listen, f.rule.Target)

	f.wg.Add(1)
	go f.acceptLoop()

	return nil
}

// Stop 停止监听并释放资源
func (f *TCPForwarder) Stop() {
	f.cancel() // 通知所有协程退出
	if f.listener != nil {
		f.listener.Close() // 关闭监听器
	}
	f.wg.Wait() // 等待所有连接处理完毕
	logging.Info("规则 %s 已停止", f.rule.Name)
}

// acceptLoop 循环接收客户端连接
func (f *TCPForwarder) acceptLoop() {
	defer f.wg.Done()

	for {
		conn, err := f.listener.Accept()
		if err != nil {
			select {
			case <-f.ctx.Done():
				// 服务正在停止，正常退出
				return
			default:
				logging.Error("规则 %s 接收连接错误: %v", f.rule.Name, err)
				continue
			}
		}

		f.wg.Add(1)
		go f.handleConnection(conn)
	}
}

// handleConnection 处理单个 TCP 连接的双向转发
func (f *TCPForwarder) handleConnection(clientConn net.Conn) {
	defer f.wg.Done()
	defer clientConn.Close()

	// 建立到目标地址的连接
	targetConn, err := net.DialTimeout("tcp", f.rule.Target, time.Duration(f.rule.TimeoutSeconds)*time.Second)
	if err != nil {
		logging.Error("规则 %s 目标地址不可达: %v", f.rule.Name, err)
		return
	}
	defer targetConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	timeout := time.Duration(f.rule.TimeoutSeconds) * time.Second

	// 客户端 -> 目标地址
	go func() {
		defer wg.Done()
		copyWithIdleTimeout(targetConn, clientConn, timeout)
		// 关闭写入端，通知对方数据发送完毕
		if tc, ok := targetConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	// 目标地址 -> 客户端
	go func() {
		defer wg.Done()
		copyWithIdleTimeout(clientConn, targetConn, timeout)
		if tc, ok := clientConn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	// 等待双向转发完成或服务停止
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-f.ctx.Done():
		// 服务停止，强制断开连接
	case <-done:
		// 正常转发结束
	}
}

// copyWithIdleTimeout 带有空闲超时的双向拷贝
func copyWithIdleTimeout(dst net.Conn, src net.Conn, timeout time.Duration) error {
	buf := make([]byte, 32*1024)
	for {
		if timeout > 0 {
			src.SetReadDeadline(time.Now().Add(timeout))
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			if timeout > 0 {
				dst.SetWriteDeadline(time.Now().Add(timeout))
			}
			nw, ew := dst.Write(buf[0:nr])
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			return nil
		}
	}
}
