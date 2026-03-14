package forward

import (
	"context"
	"net"
	"sync"
	"time"

	"portmapper/internal/config"
	"portmapper/internal/logging"
)

// UdpSession 维护一个 UDP 客户端到目标地址的会话
type UdpSession struct {
	ClientAddr     *net.UDPAddr
	TargetConn     *net.UDPConn
	LastActiveTime time.Time
}

// UDPForwarder 负责处理单条 UDP 映射规则
type UDPForwarder struct {
	rule     config.Rule
	conn     *net.UDPConn
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	sessions map[string]*UdpSession // 客户端地址 -> 会话
	mu       sync.Mutex             // 保护 sessions 映射
}

func NewUDPForwarder(rule config.Rule) *UDPForwarder {
	ctx, cancel := context.WithCancel(context.Background())
	return &UDPForwarder{
		rule:     rule,
		ctx:      ctx,
		cancel:   cancel,
		sessions: make(map[string]*UdpSession),
	}
}

// Start 启动 UDP 监听
func (f *UDPForwarder) Start() error {
	addr, err := net.ResolveUDPAddr("udp", f.rule.Listen)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	f.conn = conn

	logging.Info("规则 %s 正在监听 %s -> %s (UDP)", f.rule.Name, f.rule.Listen, f.rule.Target)

	f.wg.Add(2)
	go f.readLoop()
	go f.cleanupLoop()

	return nil
}

// Stop 停止监听并清理所有会话
func (f *UDPForwarder) Stop() {
	f.cancel()
	if f.conn != nil {
		f.conn.Close()
	}

	f.mu.Lock()
	for _, session := range f.sessions {
		session.TargetConn.Close()
	}
	f.mu.Unlock()

	f.wg.Wait()
	logging.Info("规则 %s 已停止", f.rule.Name)
}

// readLoop 循环读取客户端发来的 UDP 数据包
func (f *UDPForwarder) readLoop() {
	defer f.wg.Done()

	buf := make([]byte, 65535)
	for {
		n, clientAddr, err := f.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-f.ctx.Done():
				return
			default:
				logging.Error("规则 %s 读取错误: %v", f.rule.Name, err)
				continue
			}
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		f.handlePacket(clientAddr, data)
	}
}

// handlePacket 处理单个 UDP 数据包，建立或复用会话
func (f *UDPForwarder) handlePacket(clientAddr *net.UDPAddr, data []byte) {
	f.mu.Lock()
	session, exists := f.sessions[clientAddr.String()]
	if !exists {
		// 新客户端，创建到目标地址的连接
		targetAddr, err := net.ResolveUDPAddr("udp", f.rule.Target)
		if err != nil {
			f.mu.Unlock()
			logging.Error("规则 %s 解析目标地址错误: %v", f.rule.Name, err)
			return
		}

		targetConn, err := net.DialUDP("udp", nil, targetAddr)
		if err != nil {
			f.mu.Unlock()
			logging.Error("规则 %s 目标地址不可达: %v", f.rule.Name, err)
			return
		}

		session = &UdpSession{
			ClientAddr:     clientAddr,
			TargetConn:     targetConn,
			LastActiveTime: time.Now(),
		}
		f.sessions[clientAddr.String()] = session

		f.wg.Add(1)
		// 启动协程读取目标地址返回的数据
		go f.targetReadLoop(session, clientAddr.String())
	} else {
		// 更新活跃时间
		session.LastActiveTime = time.Now()
	}
	f.mu.Unlock()

	// 将数据转发给目标地址
	_, err := session.TargetConn.Write(data)
	if err != nil {
		logging.Error("规则 %s 写入目标地址错误: %v", f.rule.Name, err)
	}
}

// targetReadLoop 读取目标地址的返回数据，并转发回原始客户端
func (f *UDPForwarder) targetReadLoop(session *UdpSession, clientKey string) {
	defer f.wg.Done()
	defer func() {
		session.TargetConn.Close()
		f.mu.Lock()
		delete(f.sessions, clientKey)
		f.mu.Unlock()
	}()

	buf := make([]byte, 65535)
	for {
		// 设置读取超时时间，超时后自动退出协程并清理会话
		session.TargetConn.SetReadDeadline(time.Now().Add(time.Duration(f.rule.TimeoutSeconds) * time.Second))
		n, err := session.TargetConn.Read(buf)
		if err != nil {
			return
		}

		f.mu.Lock()
		session.LastActiveTime = time.Now()
		f.mu.Unlock()

		// 转发回客户端
		_, err = f.conn.WriteToUDP(buf[:n], session.ClientAddr)
		if err != nil {
			logging.Error("规则 %s 写入客户端错误: %v", f.rule.Name, err)
			return
		}
	}
}

// cleanupLoop 定时清理超时的 UDP 会话
func (f *UDPForwarder) cleanupLoop() {
	defer f.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.Duration(f.rule.TimeoutSeconds) * time.Second

	for {
		select {
		case <-f.ctx.Done():
			return
		case now := <-ticker.C:
			f.mu.Lock()
			for key, session := range f.sessions {
				if now.Sub(session.LastActiveTime) > timeout {
					session.TargetConn.Close()
					delete(f.sessions, key)
				}
			}
			f.mu.Unlock()
		}
	}
}
