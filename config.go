package main

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Lyxot/CloudflareSpeedTestDNS/ddns"
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
	IPv4File string `toml:"ipv4_file"` // IPv4段数据文件
	IPv6File string `toml:"ipv6_file"` // IPv6段数据文件
	IPText   string `toml:"ip_text"`   // 指定IP段数据
	Output   string `toml:"output"`    // 输出结果文件
	LogFile  string `toml:"log_file"`  // 日志文件

	// 其他选项
	TestAll bool `toml:"test_all"` // 测速全部IP
	Debug   bool `toml:"debug"`   // 调试输出模式

	// 阿里云DNS相关
	AliDNS AliDNSConfig `toml:"alidns"` // 阿里云DNS配置

	// DNSPod DNS相关
	DNSPod DNSPodConfig `toml:"dnspod"` // DNSPod DNS配置

	// Cloudflare DNS相关
	Cloudflare CloudflareConfig `toml:"cloudflare"` // Cloudflare DNS配置

	// Cloudflare KV相关
	CFKV CloudflareKVConfig `toml:"cfkv"` // Cloudflare KV配置

	// Cron 定时任务相关
	Cron CronConfig `toml:"cron"`
}

// AliDNSConfig 阿里云DNS配置
type AliDNSConfig struct {
	Enable          bool   `toml:"enable"`           // 是否启用阿里云DNS
	AccessKeyID     string `toml:"accesskey_id"`     // 阿里云AccessKeyID
	AccessKeySecret string `toml:"accesskey_secret"` // 阿里云AccessKeySecret
	Domain          string `toml:"domain"`           // 域名
	Subdomain       string `toml:"subdomain"`        // 子域名
	TTL             int    `toml:"ttl"`              // TTL
}

// DNSPodConfig DNSPod DNS配置
type DNSPodConfig struct {
	Enable    bool   `toml:"enable"`     // 是否启用DNSPod DNS
	SecretID  string `toml:"secret_id"`  // DNSPod Secret ID
	SecretKey string `toml:"secret_key"` // DNSPod Secret Key
	Domain    string `toml:"domain"`     // 域名
	Subdomain string `toml:"subdomain"`  // 子域名
	TTL       int    `toml:"ttl"`        // TTL
}

// CloudflareConfig Cloudflare DNS配置
type CloudflareConfig struct {
	Enable    bool   `toml:"enable"`    // 是否启用Cloudflare DNS
	APIToken  string `toml:"api_token"` // Cloudflare API Token
	ZoneID    string `toml:"zone_id"`   // Cloudflare Zone ID
	Domain    string `toml:"domain"`    // 域名
	Subdomain string `toml:"subdomain"` // 子域名
	Proxied   bool   `toml:"proxied"`   // 是否开启Cloudflare代理
	TTL       int    `toml:"ttl"`       // TTL，1为自动
}

// CloudflareKVConfig Cloudflare KV配置
type CloudflareKVConfig struct {
	Enable      bool   `toml:"enable"`       // 是否启用Cloudflare KV
	APIToken    string `toml:"api_token"`    // Cloudflare API Token
	AccountID   string `toml:"account_id"`   // Cloudflare Account ID
	NamespaceID string `toml:"namespace_id"` // Cloudflare KV Namespace ID
}

// CronConfig Cron 定时任务相关参数
type CronConfig struct {
	Enable            bool    `toml:"enable"`
	LatencyThreshold  int     `toml:"latency_threshold"`
	LossRateThreshold float64 `toml:"loss_rate_threshold"`
	CheckInterval     int     `toml:"check_interval"`
	TestInterval      int     `toml:"test_interval"`
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

// CreateDefaultConfig 创建默认配置
func CreateDefaultConfig() *Config {
	return &Config{
		Routines:     200,
		PingTimes:    4,
		TCPPort:      443,
		MaxDelay:     9999,
		MinDelay:     0,
		MaxLossRate:  1.0,
		Httping:      false,
		HttpingCode:  0,
		CFColo:       "",
		TestCount:    10,
		DownloadTime: 10,
		URL:          "https://cf.xiu2.xyz/url",
		MinSpeed:     0.0,
		Disable:      false,
		PrintNum:     10,
		IPFile:       "ip.txt",
		IPv4File:     "",
		IPv6File:     "",
		IPText:       "",
		Output:       "result.csv",
		LogFile:      "",
		TestAll:      false,
		AliDNS: AliDNSConfig{
			Enable: false,
			TTL:    600,
		},
		DNSPod: DNSPodConfig{
			Enable: false,
			TTL:    600,
		},
		Cloudflare: CloudflareConfig{
			Enable:  false,
			Proxied: false,
			TTL:     1,
		},
		CFKV: CloudflareKVConfig{
			Enable: false,
		},
		Cron: CronConfig{
			Enable:            false,
			LatencyThreshold:  9999,
			LossRateThreshold: 1.0,
			CheckInterval:     30,
			TestInterval:      24,
		},
	}
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

	if config.IPv4File != "" {
		task.IPv4File = config.IPv4File
	}

	if config.IPv6File != "" {
		task.IPv6File = config.IPv6File
	}

	if config.IPText != "" {
		task.IPText = config.IPText
	}

	// 智能判断文件优先级
	if config.IPv4File != "" || config.IPv6File != "" {
		// 如果指定了ipv4_file或ipv6_file，则ip_file参数无效
		task.IPFile = ""
	}

	// 设置其他选项
	task.TestAll = config.TestAll
	utils.Debug = config.Debug

	// 设置阿里云DNS相关参数
	enableAliDNS = config.AliDNS.Enable
	ddns.AliDNSConfig.AccessKeyID = config.AliDNS.AccessKeyID
	ddns.AliDNSConfig.AccessKeySecret = config.AliDNS.AccessKeySecret
	ddns.AliDNSConfig.Domain = config.AliDNS.Domain
	ddns.AliDNSConfig.Subdomain = config.AliDNS.Subdomain
	ddns.AliDNSConfig.TTL = config.AliDNS.TTL

	// 设置DNSPod DNS相关参数
	enableDNSPod = config.DNSPod.Enable
	ddns.DNSPodConfig.SecretID = config.DNSPod.SecretID
	ddns.DNSPodConfig.SecretKey = config.DNSPod.SecretKey
	ddns.DNSPodConfig.Domain = config.DNSPod.Domain
	ddns.DNSPodConfig.Subdomain = config.DNSPod.Subdomain
	ddns.DNSPodConfig.TTL = config.DNSPod.TTL

	// 设置Cloudflare DNS相关参数
	enableCloudflare = config.Cloudflare.Enable
	ddns.CloudflareConfig.APIToken = config.Cloudflare.APIToken
	ddns.CloudflareConfig.ZoneID = config.Cloudflare.ZoneID
	ddns.CloudflareConfig.Domain = config.Cloudflare.Domain
	ddns.CloudflareConfig.Subdomain = config.Cloudflare.Subdomain
	ddns.CloudflareConfig.Proxied = config.Cloudflare.Proxied
	ddns.CloudflareConfig.TTL = config.Cloudflare.TTL

	// 设置Cloudflare KV相关参数
	enableCFKV = config.CFKV.Enable
	ddns.CloudflareKVConfig.APIToken = config.CFKV.APIToken
	ddns.CloudflareKVConfig.AccountID = config.CFKV.AccountID
	ddns.CloudflareKVConfig.NamespaceID = config.CFKV.NamespaceID

	// 设置输入输出相关参数
	if config.PrintNum >= 0 {
		utils.PrintNum = config.PrintNum
	}

	if config.Output != "" {
		utils.Output = config.Output
	}

	if config.LogFile != "" {
		utils.LogFile = config.LogFile
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

	// 设置Cron定时任务相关参数
	if config.Cron.Enable {
		enableCron = true
		if config.Cron.LatencyThreshold > 0 {
			latencyThreshold = time.Duration(config.Cron.LatencyThreshold) * time.Millisecond
		}
		if config.Cron.LossRateThreshold >= 0 && config.Cron.LossRateThreshold <= 1 {
			lossRateThreshold = float32(config.Cron.LossRateThreshold)
		}
		if config.Cron.CheckInterval > 0 {
			checkInterval = time.Duration(config.Cron.CheckInterval) * time.Minute
		}
		if config.Cron.TestInterval > 0 {
			testInterval = time.Duration(config.Cron.TestInterval) * time.Hour
		}
	}
}
