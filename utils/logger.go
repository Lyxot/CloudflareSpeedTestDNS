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
	LOG_DEBUG = "DEBUG"
	LOG_INFO  = "INFO"
	LOG_WARN  = "WARN"
	LOG_ERROR = "ERROR"
	LOG_FATAL = "FATAL"
)

// LogInfo 输出信息日志
func LogInfo(format string, args ...interface{}) {
	logMessage(LOG_INFO, format, args...)
}

// LogError 输出错误日志
func LogError(format string, args ...interface{}) {
	logMessage(LOG_ERROR, format, args...)
}

// LogWarn 输出警告日志
func LogWarn(format string, args ...interface{}) {
	logMessage(LOG_WARN, format, args...)
}

// LogDebug 输出调试日志
func LogDebug(format string, args ...interface{}) {
	logMessage(LOG_DEBUG, format, args...)
}

// LogFatal 输出致命错误日志并退出
func LogFatal(format string, args ...interface{}) {
	logMessage(LOG_FATAL, format, args...)
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
		logFile.Close()
		InitLogFile()
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
	case LOG_DEBUG:
		levelColor = Yellow
	case LOG_INFO:
		levelColor = White
	case LOG_WARN:
		levelColor = Red
		output = os.Stderr
	case LOG_ERROR:
		levelColor = Red
		output = os.Stderr
	case LOG_FATAL:
		levelColor = Red
		output = os.Stderr
	default:
		levelColor = White
	}

	// 格式化日志消息
	message := fmt.Sprintf(format, args...)

	// 输出到屏幕
	Green.Fprintf(output, "%s ", timestamp)
	levelColor.Fprintf(output, "%s", level)
	fmt.Fprintf(output, " %s\n", message)

	// 如果配置了日志文件，同时输出到文件
	if LogFile != "" {
		logContent := fmt.Sprintf("%s %s %s\n", timestamp, level, message)
		writeToLogFile(logContent)
	}
}
