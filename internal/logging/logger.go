package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

// dailyFileWriter 实现按天轮转的日志写入器
type dailyFileWriter struct {
	mu       sync.Mutex
	basePath string
	file     *os.File
	currDate string
}

func newDailyFileWriter(basePath string) (*dailyFileWriter, error) {
	w := &dailyFileWriter{basePath: basePath}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	// 启动时清理一次旧日志
	go w.cleanOldLogs()
	return w, nil
}

func (w *dailyFileWriter) rotate() error {
	nowDate := time.Now().Format("2006-01-02")
	if w.currDate == nowDate && w.file != nil {
		return nil
	}

	if w.file != nil {
		w.file.Close()
	}

	dir := filepath.Dir(w.basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	ext := filepath.Ext(w.basePath)
	name := w.basePath[:len(w.basePath)-len(ext)]
	logPath := fmt.Sprintf("%s-%s%s", name, nowDate, ext)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	w.file = file
	w.currDate = nowDate

	// 每次轮转时清理旧日志
	go w.cleanOldLogs()

	return nil
}

// cleanOldLogs 清理超过 15 天的旧日志
func (w *dailyFileWriter) cleanOldLogs() {
	dir := filepath.Dir(w.basePath)
	ext := filepath.Ext(w.basePath)
	baseName := filepath.Base(w.basePath)
	namePrefix := baseName[:len(baseName)-len(ext)] + "-"

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -15)

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileName := f.Name()
		if len(fileName) > len(namePrefix) && fileName[:len(namePrefix)] == namePrefix && filepath.Ext(fileName) == ext {
			dateStr := fileName[len(namePrefix) : len(fileName)-len(ext)]
			if fileDate, err := time.Parse("2006-01-02", dateStr); err == nil {
				if fileDate.Before(cutoff) {
					os.Remove(filepath.Join(dir, fileName))
				}
			}
		}
	}
}

func (w *dailyFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotate(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

// customWriter 自定义日志输出格式，包含时间戳和日志级别
type customWriter struct {
	out   io.Writer
	level string
}

func (w *customWriter) Write(p []byte) (n int, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	// log.Printf 已经自带换行符，因此这里不需要再加 \n
	msg := fmt.Sprintf("%s [%s] %s", timestamp, w.level, string(p))
	return w.out.Write([]byte(msg))
}

// InitLogger 初始化全局日志记录器
// isService: 是否以服务模式运行。服务模式下日志写入文件，前台模式写入控制台
func InitLogger(logPath string, isService bool) error {
	var out io.Writer = os.Stdout

	if isService && logPath != "" {
		// 如果是相对路径，转换为基于可执行文件所在目录的绝对路径
		// 防止 Windows 服务在 C:\Windows\System32 下生成日志
		if !filepath.IsAbs(logPath) {
			exePath, err := os.Executable()
			if err == nil {
				logPath = filepath.Join(filepath.Dir(exePath), logPath)
			}
		}

		fw, err := newDailyFileWriter(logPath)
		if err != nil {
			return err
		}
		out = fw
	}

	InfoLogger = log.New(&customWriter{out: out, level: "INFO"}, "", 0)
	WarnLogger = log.New(&customWriter{out: out, level: "WARN"}, "", 0)
	ErrorLogger = log.New(&customWriter{out: out, level: "ERROR"}, "", 0)

	return nil
}

func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if WarnLogger != nil {
		WarnLogger.Printf(format, v...)
	}
}

func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}
