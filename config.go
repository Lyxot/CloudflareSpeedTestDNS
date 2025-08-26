package main

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Lyxot/CloudflareSpeedTestDNS/task"
	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
)

// Config 配置文件结构体
type Config struct {
	// 延迟测速相关
	Routines    int     `toml:"routines"`      // 延迟测速线程
	PingTimes   int     `toml:"ping_times"`    // 延迟测速次数
	TCPPort     int     `toml:"tcp_port"`      // 指定测速端口
	MaxDelay    int     `toml:"max_delay"`     // 平均延迟上限
	MinDelay    int     `toml:"min_delay"`     // 平均延迟下限
	MaxLossRate float64 `toml:"max_loss_rate"` // 丢包几率上限

	// HTTP测速相关
	Httping     bool   `toml:"httping"`      // 切换测速模式为HTTP
	HttpingCode int    `toml:"httping_code"` // 有效状态代码
	CFColo      string `toml:"cfcolo"`       // 匹配指定地区

	// 下载测速相关
	TestCount    int     `toml:"test_count"`       // 下载测速数量
	DownloadTime int     `toml:"download_time"`    // 下载测速时间
	URL          string  `toml:"url"`              // 指定测速地址
	MinSpeed     float64 `toml:"min_speed"`        // 下载速度下限
	Disable      bool    `toml:"disable_download"` // 禁用下载测速

	// 输入输出相关
	PrintNum int    `toml:"print_num"` // 显示结果数量
	IPFile   string `toml:"ip_file"`   // IP段数据文件
	IPText   string `toml:"ip_text"`   // 指定IP段数据
	Output   string `toml:"output"`    // 输出结果文件

	// 其他选项
	TestAll bool `toml:"test_all"` // 测速全部IP
	Debug   bool `toml:"debug"`    // 调试输出模式
}

// LoadConfig 从TOML文件加载配置
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	// 检查文件是否存在
	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("配置文件 %s 不存在或无法访问: %v", path, err)
	}

	// 解析TOML文件
	_, err = toml.DecodeFile(path, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// ApplyConfig 应用配置到全局变量
func ApplyConfig(config *Config) {
	// 设置延迟测速相关参数
	if config.Routines > 0 {
		task.Routines = config.Routines
	}

	if config.PingTimes > 0 {
		task.PingTimes = config.PingTimes
	}

	if config.TCPPort > 0 {
		task.TCPPort = config.TCPPort
	}

	// 设置HTTP测速相关参数
	task.Httping = config.Httping

	if config.HttpingCode > 0 {
		task.HttpingStatusCode = config.HttpingCode
	}

	if config.CFColo != "" {
		task.HttpingCFColo = config.CFColo
		task.HttpingCFColomap = task.MapColoMap()
	}

	// 设置下载测速相关参数
	if config.TestCount > 0 {
		task.TestCount = config.TestCount
	}

	if config.DownloadTime > 0 {
		task.Timeout = time.Duration(config.DownloadTime) * time.Second
	}

	if config.URL != "" {
		task.URL = config.URL
	}

	if config.MinSpeed >= 0 {
		task.MinSpeed = config.MinSpeed
	}

	task.Disable = config.Disable

	// 设置输入输出相关参数
	if config.IPFile != "" {
		task.IPFile = config.IPFile
	}

	if config.IPText != "" {
		task.IPText = config.IPText
	}

	// 设置其他选项
	task.TestAll = config.TestAll

	// 设置输入输出相关参数
	if config.PrintNum >= 0 {
		utils.PrintNum = config.PrintNum
	}

	if config.Debug {
		utils.Debug = config.Debug
	}

	if config.Output != "" {
		utils.Output = config.Output
	}

	// 设置延迟相关参数
	if config.MaxDelay > 0 {
		utils.InputMaxDelay = time.Duration(config.MaxDelay) * time.Millisecond
	}

	if config.MinDelay >= 0 {
		utils.InputMinDelay = time.Duration(config.MinDelay) * time.Millisecond
	}

	if config.MaxLossRate >= 0 && config.MaxLossRate <= 1 {
		utils.InputMaxLossRate = float32(config.MaxLossRate)
	}
}
