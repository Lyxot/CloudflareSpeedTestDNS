package conf

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Lyxot/CloudflareSpeedTestDNS/ddns"
	"github.com/Lyxot/CloudflareSpeedTestDNS/task"
	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
)

var (
	EnableAliDNS      bool
	EnableCloudflare  bool
	EnableCFKV        bool
	EnableDNSPod      bool
	EnableCron        bool
	LatencyThreshold  time.Duration
	LossRateThreshold float32
	CheckInterval     time.Duration
	TestInterval      time.Duration
)

// Config 配置文件结构体
type Config struct {
	// 延迟测速相关
	Routines    int     `toml:"routines"`      // 延迟测速线程
	PingTimes   int     `toml:"ping_times"`    // 延迟测速次数
	TcpPort     int     `toml:"tcp_port"`      // 指定测速端口
	MaxDelay    int     `toml:"max_delay"`     // 平均延迟上限
	MinDelay    int     `toml:"min_delay"`     // 平均延迟下限
	MaxLossRate float64 `toml:"max_loss_rate"` // 丢包几率上限

	// HTTP测速相关
	Httping     bool   `toml:"httping"`      // 切换测速模式为HTTP
	HttpingCode int    `toml:"httping_code"` // 有效状态代码
	Cfcolo      string `toml:"cfcolo"`       // 匹配指定地区

	// 下载测速相关
	TestCount       int     `toml:"test_count"`       // 下载测速数量
	DownloadTime    int     `toml:"download_time"`    // 下载测速时间
	Url             string  `toml:"url"`              // 指定测速地址
	MinSpeed        float64 `toml:"min_speed"`        // 下载速度下限
	DisableDownload bool    `toml:"disable_download"` // 禁用下载测速

	// 输入输出相关
	PrintNum int    `toml:"print_num"` // 显示结果数量
	IpFile   string `toml:"ip_file"`   // IP段数据文件
	Ipv4File string `toml:"ipv4_file"` // IPv4段数据文件
	Ipv6File string `toml:"ipv6_file"` // IPv6段数据文件
	IpText   string `toml:"ip_text"`   // 指定IP段数据
	Output   string `toml:"output"`    // 输出结果文件
	LogFile  string `toml:"log_file"`  // 日志文件

	// 其他选项
	TestAll bool `toml:"test_all"` // 测速全部IP
	Debug   bool `toml:"debug"`    // 调试输出模式

	// 阿里云DNS相关
	Alidns AliDNSConfig `toml:"alidns"` // 阿里云DNS配置

	// Dnspod DNS相关
	Dnspod DNSPodConfig `toml:"dnspod"` // DNSPod DNS配置

	// Cloudflare DNS相关
	Cloudflare CloudflareConfig `toml:"cloudflare"` // Cloudflare DNS配置

	// Cloudflare KV相关
	Cfkv CloudflareKVConfig `toml:"cfkv"` // Cloudflare KV配置

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
		Routines:        200,
		PingTimes:       4,
		TcpPort:         443,
		MaxDelay:        9999,
		MinDelay:        0,
		MaxLossRate:     1.0,
		Httping:         false,
		HttpingCode:     0,
		Cfcolo:          "",
		TestCount:       10,
		DownloadTime:    10,
		Url:             "https://cf.xiu2.xyz/url",
		MinSpeed:        0.0,
		DisableDownload: false,
		PrintNum:        10,
		IpFile:          "ip.txt",
		Ipv4File:        "",
		Ipv6File:        "",
		IpText:          "",
		Output:          "result.csv",
		LogFile:         "",
		TestAll:         false,
		Alidns: AliDNSConfig{
			Enable: false,
			TTL:    600,
		},
		Dnspod: DNSPodConfig{
			Enable: false,
			TTL:    600,
		},
		Cloudflare: CloudflareConfig{
			Enable:  false,
			Proxied: false,
			TTL:     1,
		},
		Cfkv: CloudflareKVConfig{
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

	if config.TcpPort > 0 {
		task.TCPPort = config.TcpPort
	}

	// 设置HTTP测速相关参数
	task.Httping = config.Httping

	if config.HttpingCode > 0 {
		task.HttpingStatusCode = config.HttpingCode
	}

	if config.Cfcolo != "" {
		task.HttpingCFColo = config.Cfcolo
		task.HttpingCFColoMap = task.MapColoMap()
	}

	// 设置下载测速相关参数
	if config.TestCount > 0 {
		task.TestCount = config.TestCount
	}

	if config.DownloadTime > 0 {
		task.Timeout = time.Duration(config.DownloadTime) * time.Second
	}

	if config.Url != "" {
		task.URL = config.Url
	}

	if config.MinSpeed >= 0 {
		task.MinSpeed = config.MinSpeed
	}

	task.Disable = config.DisableDownload

	// 设置输入输出相关参数
	if config.IpFile != "" {
		task.IPFile = config.IpFile
	}

	if config.Ipv4File != "" {
		task.IPv4File = config.Ipv4File
	}

	if config.Ipv6File != "" {
		task.IPv6File = config.Ipv6File
	}

	if config.IpText != "" {
		task.IPText = config.IpText
	}

	// 智能判断文件优先级
	if config.Ipv4File != "" || config.Ipv6File != "" {
		// 如果指定了ipv4_file或ipv6_file，则ip_file参数无效
		task.IPFile = ""
	}

	// 设置其他选项
	task.TestAll = config.TestAll
	utils.Debug = config.Debug

	// 设置阿里云DNS相关参数
	EnableAliDNS = config.Alidns.Enable
	ddns.AliDNSConfig.AccessKeyID = config.Alidns.AccessKeyID
	ddns.AliDNSConfig.AccessKeySecret = config.Alidns.AccessKeySecret
	ddns.AliDNSConfig.Domain = config.Alidns.Domain
	ddns.AliDNSConfig.Subdomain = config.Alidns.Subdomain
	ddns.AliDNSConfig.TTL = config.Alidns.TTL

	// 设置DNSPod DNS相关参数
	EnableDNSPod = config.Dnspod.Enable
	ddns.DNSPodConfig.SecretID = config.Dnspod.SecretID
	ddns.DNSPodConfig.SecretKey = config.Dnspod.SecretKey
	ddns.DNSPodConfig.Domain = config.Dnspod.Domain
	ddns.DNSPodConfig.Subdomain = config.Dnspod.Subdomain
	ddns.DNSPodConfig.TTL = config.Dnspod.TTL

	// 设置Cloudflare DNS相关参数
	EnableCloudflare = config.Cloudflare.Enable
	ddns.CloudflareConfig.APIToken = config.Cloudflare.APIToken
	ddns.CloudflareConfig.ZoneID = config.Cloudflare.ZoneID
	ddns.CloudflareConfig.Domain = config.Cloudflare.Domain
	ddns.CloudflareConfig.Subdomain = config.Cloudflare.Subdomain
	ddns.CloudflareConfig.Proxied = config.Cloudflare.Proxied
	ddns.CloudflareConfig.TTL = config.Cloudflare.TTL

	// 设置Cloudflare KV相关参数
	EnableCFKV = config.Cfkv.Enable
	ddns.CloudflareKVConfig.APIToken = config.Cfkv.APIToken
	ddns.CloudflareKVConfig.AccountID = config.Cfkv.AccountID
	ddns.CloudflareKVConfig.NamespaceID = config.Cfkv.NamespaceID

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
		EnableCron = true
		if config.Cron.LatencyThreshold > 0 {
			LatencyThreshold = time.Duration(config.Cron.LatencyThreshold) * time.Millisecond
		}
		if config.Cron.LossRateThreshold >= 0 && config.Cron.LossRateThreshold <= 1 {
			LossRateThreshold = float32(config.Cron.LossRateThreshold)
		}
		if config.Cron.CheckInterval > 0 {
			CheckInterval = time.Duration(config.Cron.CheckInterval) * time.Minute
		}
		if config.Cron.TestInterval > 0 {
			TestInterval = time.Duration(config.Cron.TestInterval) * time.Hour
		}
	}
}
