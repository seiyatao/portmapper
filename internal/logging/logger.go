package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

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
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		out = file
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
