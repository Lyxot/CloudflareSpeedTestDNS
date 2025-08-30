package utils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// 日志级别常量
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelFatal = "FATAL"
)

// LogInfo 输出信息日志
func LogInfo(format string, args ...interface{}) {
	logMessage(LogLevelInfo, format, args...)
}

// LogError 输出错误日志
func LogError(format string, args ...interface{}) {
	logMessage(LogLevelError, format, args...)
}

// LogWarn 输出警告日志
func LogWarn(format string, args ...interface{}) {
	logMessage(LogLevelWarn, format, args...)
}

// LogDebug 输出调试日志
func LogDebug(format string, args ...interface{}) {
	logMessage(LogLevelDebug, format, args...)
}

// LogFatal 输出致命错误日志并退出
func LogFatal(format string, args ...interface{}) {
	logMessage(LogLevelFatal, format, args...)
	os.Exit(1)
}

var (
	LogFile  = "" // 日志文件路径
	logFile  *os.File
	logMutex sync.Mutex
)

// InitLogFile 初始化日志文件
func InitLogFile() error {
	if LogFile == "" {
		return nil
	}

	var err error
	logFile, err = os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return err
}

// writeToLogFile 写入日志文件
func writeToLogFile(content string) {
	if logFile == nil {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	_, err := logFile.WriteString(content)
	if err != nil {
		// 如果写入失败，尝试重新打开文件
		err = logFile.Close()
		err = InitLogFile()
		if err != nil {
			return
		}
	}
}

// logMessage 统一的日志输出函数
func logMessage(level string, format string, args ...interface{}) {
	// 获取当前时间并格式化为 hh:mm:ss
	timestamp := time.Now().Format("15:04:05")

	// 根据日志级别选择颜色
	var levelColor *color.Color
	output := os.Stdout
	switch level {
	case LogLevelDebug:
		levelColor = Yellow
	case LogLevelInfo:
		levelColor = White
	case LogLevelWarn:
		levelColor = Red
		output = os.Stderr
	case LogLevelError:
		levelColor = Red
		output = os.Stderr
	case LogLevelFatal:
		levelColor = Red
		output = os.Stderr
	default:
		levelColor = White
	}

	// 格式化日志消息
	message := fmt.Sprintf(format, args...)

	// 输出到屏幕
	_, err := Green.Fprintf(output, "%s ", timestamp)
	_, err = levelColor.Fprintf(output, "%s", level)
	_, err = fmt.Fprintf(output, " %s\n", message)

	if err != nil {
		// 如果输出失败，直接返回
		return
	}

	// 如果配置了日志文件，同时输出到文件
	if LogFile != "" {
		logContent := fmt.Sprintf("%s %s %s\n", timestamp, level, message)
		writeToLogFile(logContent)
	}
}
