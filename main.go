package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/Lyxot/CloudflareSpeedTestDNS/conf"
	"github.com/Lyxot/CloudflareSpeedTestDNS/ddns"
	"github.com/Lyxot/CloudflareSpeedTestDNS/task"
	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
)

var (
	version    string
	gitCommit  string
	configFile string
)

func init() {
	var printVersion, checkUpdateFlag, debugFlag, pgoFlag bool
	var help = `CloudflareSpeedTestDNS ` + version + `-` + gitCommit + `
测试各个 CDN 或网站所有 IP 的延迟和速度，获取最快 IP (IPv4+IPv6)！
https://github.com/Lyxot/CloudflareSpeedTestDNS

参数：
    -c config.toml
        指定TOML配置文件；默认为config.toml，不存在时使用默认参数
    -debug
        调试输出模式；会在一些非预期情况下输出更多日志以便判断原因；(默认 关闭)
	-pgo
		开启 CPU 性能分析
    -v
        打印程序版本
    -u
        检查版本更新
    -h
        打印帮助说明
`
	flag.BoolVar(&debugFlag, "debug", false, "调试输出模式")
	flag.BoolVar(&pgoFlag, "pgo", false, "开启 CPU 性能分析")
	flag.StringVar(&configFile, "c", "", "指定TOML配置文件")
	flag.BoolVar(&printVersion, "v", false, "打印程序版本")
	flag.BoolVar(&checkUpdateFlag, "u", false, "检查版本更新")
	flag.Usage = func() { fmt.Print(help) }
	flag.Parse()

	if pgoFlag {
		pgo()
	}

	if printVersion {
		fmt.Printf("CloudflareSpeedTestDNS version %s, build %s, %s\n", version, gitCommit, runtime.Version())
		endPrint()
		os.Exit(0)
	}

	if checkUpdateFlag {
		fmt.Println("检查版本更新中...")
		versionNew, err := checkUpdate()
		if err != nil {
			_, _ = utils.Red.Printf("检查版本更新失败: %v", err)
		} else if versionNew != "" && versionNew != version {
			_, _ = utils.Yellow.Printf("*** 发现新版本 [%s]！请前往 [https://github.com/Lyxot/CloudflareSpeedTestDNS/releases/latest] 更新！ ***", versionNew)
		} else {
			_, _ = utils.Green.Println("当前为最新版本 [" + version + "]！")
		}
		fmt.Printf("\n")
		endPrint()
		os.Exit(0)
	}

	var config *conf.Config
	var err error

	if configFile != "" {
		// 如果指定了配置文件，则加载它
		config, err = conf.LoadConfig(configFile)
		if err != nil {
			utils.LogFatal("加载配置文件失败: %v", err)
		}
	} else {
		// 如果未指定配置文件，则尝试加载默认的 config.toml
		config, err = conf.LoadConfig("config.toml")
		if err != nil {
			utils.LogWarn("加载配置文件 [config.toml] 失败: %v，改用默认配置", err)
			config = conf.CreateDefaultConfig()
		}
	}

	conf.LoadEnvConfig(config)
	conf.ApplyConfig(config)

	// 如果通过命令行指定了 -debug，则覆盖配置文件中的设置
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "debug" {
			utils.Debug = debugFlag
		}
	})

	// 初始化日志文件
	if err := utils.InitLogFile(); err != nil {
		utils.LogFatal("初始化日志文件失败: %v", err)
	}

	if task.MinSpeed > 0 && config.MaxDelay == 9999 {
		utils.LogWarn("配置了 min_speed 参数时，建议搭配 max_delay 参数，以避免因凑不够 test_count 数量而一直测速...")
	}
}

func main() {
	utils.LogInfo("# Lyxot/CloudflareSpeedTestDNS %s-%s", version, gitCommit)
	if conf.EnableCron {
		cron() // 定时任务
	} else {
		speedTest() // 开始测速
	}
	endPrint() // 根据情况选择退出方式（针对 Windows）
}

func cron() {
	utils.LogInfo("定时任务已启用")
	ipData := speedTest()

	// 设置定时器
	testTicker := time.NewTicker(conf.TestInterval)
	checkTicker := time.NewTicker(conf.CheckInterval)

	for {
		select {
		case <-testTicker.C:
			utils.LogInfo("强制刷新任务开始...")
			ipData = speedTest()
			checkTicker.Reset(conf.CheckInterval)
		case <-checkTicker.C:
			utils.LogInfo("开始检查延迟和丢包率...")
			// 拼接 IP 段数据
			ipText := ""
			for _, ip := range ipData {
				ipText += ip + ","
			}
			// 保存原始设置
			origIPText := task.IPText
			origMaxDelay := utils.InputMaxDelay
			origMaxLossRate := utils.InputMaxLossRate

			task.IPText = ipText
			utils.InputMaxDelay = conf.LatencyThreshold
			utils.InputMaxLossRate = conf.LossRateThreshold
			pingData := task.NewPing().Run().FilterDelay().FilterLossRate()

			// 恢复原始设置
			task.IPText = origIPText
			utils.InputMaxDelay = origMaxDelay
			utils.InputMaxLossRate = origMaxLossRate

			if len(pingData) != len(ipData) {
				utils.LogInfo("延迟或丢包率超过阈值，开始新一轮测速...")
				ipData = speedTest()
				testTicker.Reset(conf.TestInterval)
			} else {
				utils.LogInfo("延迟和丢包率在阈值范围内")
			}
		}
	}
}

func speedTest() []string {
	var ipData []string
	if task.IsBothMode() {
		// 保存原始文件设置
		origIPv4File := task.IPv4File
		origIPv6File := task.IPv6File
		originOutput := utils.Output

		// 测试IPv4
		utils.LogInfo("[IPv4] 开始测试IPv4...")
		task.IPv6File = ""
		utils.Output = utils.GetFilenameWithSuffix(originOutput, "ipv4")
		ipv4SpeedData := singleSpeedTest()                  // 开始延迟测速 + 过滤延迟/丢包
		ipData = append(ipData, ddnsSync(ipv4SpeedData)...) // 同步到DNS

		// 测试IPv6
		utils.LogInfo("[IPv6] 开始测试IPv6...")
		task.IPv4File = ""
		task.IPv6File = origIPv6File
		utils.Output = utils.GetFilenameWithSuffix(originOutput, "ipv6")
		ipv6SpeedData := singleSpeedTest()                  // 开始延迟测速 + 过滤延迟/丢包
		ipData = append(ipData, ddnsSync(ipv6SpeedData)...) // 同步到DNS

		// 恢复原始文件设置
		task.IPv4File = origIPv4File
		task.IPv6File = origIPv6File
		utils.Output = originOutput
	} else {
		ipData = ddnsSync(singleSpeedTest()) // 延迟测速 + 过滤延迟/丢包 + 同步到DNS
	}
	return ipData
}

func singleSpeedTest() utils.DownloadSpeedSet {
	var speedData utils.DownloadSpeedSet
	for i := 0; i < conf.MaxAttempts; i++ {
		// 开始延迟测速 + 过滤延迟/丢包
		pingData := task.NewPing().Run().FilterDelay().FilterLossRate()
		// 开始下载测速
		speedData = task.TestDownloadSpeed(pingData)
		if len(speedData) >= conf.MinNum {
			break
		}
		if i < conf.MaxAttempts-1 {
			utils.LogWarn("符合条件的IP数量[%d]少于设定的最小数量[%d]，将在3秒后开始新一轮测试...", len(speedData), conf.MinNum)
			time.Sleep(3 * time.Second)
		} else {
			utils.LogWarn("符合条件的IP数量[%d]少于设定的最小数量[%d]，已达到最大重试次数，测试结束。", len(speedData), conf.MinNum)
		}
	}
	utils.ExportCsv(speedData) // 输出文件
	speedData.Print()          // 打印结果

	return speedData
}

func ddnsSync(speedData utils.DownloadSpeedSet) []string {
	if len(speedData) == 0 {
		return []string{}
	}

	// 根据结果类型分类
	var ipv4Results []string
	var ipv6Results []string
	for i := 0; i < utils.PrintNum && i < len(speedData); i++ {
		ip := speedData[i].IP.String()
		if task.IsIPv4(ip) {
			ipv4Results = append(ipv4Results, ip)
		} else {
			ipv6Results = append(ipv6Results, ip)
		}
	}

	// 如果启用了阿里云DNS，则同步结果
	if conf.EnableAliDNS {
		utils.LogInfo("开始同步结果到阿里云DNS...")
		if err := ddns.SyncDNSRecords(ipv4Results, ipv6Results); err != nil {
			utils.LogError("同步到阿里云DNS失败: %v", err)
		} else {
			utils.LogInfo("同步到阿里云DNS成功!")
		}
	}

	// 如果启用了DNSPod DNS，则同步结果
	if conf.EnableDNSPod {
		utils.LogInfo("开始同步结果到DNSPod DNS...")
		if err := ddns.SyncDNSPodRecords(ipv4Results, ipv6Results); err != nil {
			utils.LogError("同步到DNSPod DNS失败: %v", err)
		} else {
			utils.LogInfo("同步到DNSPod DNS成功!")
		}
	}

	// 如果启用了Cloudflare DNS，则同步结果
	if conf.EnableCloudflare {
		utils.LogInfo("开始同步结果到Cloudflare DNS...")
		if err := ddns.SyncCloudflareRecords(ipv4Results, ipv6Results); err != nil {
			utils.LogError("同步到Cloudflare DNS失败: %v", err)
		} else {
			utils.LogInfo("同步到Cloudflare DNS成功!")
		}
	}

	// 如果启用了Cloudflare KV，则同步结果
	if conf.EnableCFKV {
		utils.LogInfo("开始同步结果到Cloudflare KV...")
		if err := ddns.SyncCloudflareKV(speedData.FilterIPv4(), speedData.FilterIPv6()); err != nil {
			utils.LogError("同步到Cloudflare KV失败: %v", err)
		} else {
			utils.LogInfo("同步到Cloudflare KV成功!")
		}
	}

	return append(ipv4Results, ipv6Results...)
}

// 根据情况选择退出方式（针对 Windows）
func endPrint() {
	if utils.NoPrintResult() { // 如果不需要打印测速结果，则直接退出
		return
	}
	if runtime.GOOS == "windows" { // 如果是 Windows 系统，则需要按下 回车键 或 Ctrl+C 退出（避免通过双击运行时，测速完毕后直接关闭）
		fmt.Println("按下 回车键 或 Ctrl+C 退出。")
		fmt.Scanln()
	}
}

// 检查更新
func checkUpdate() (string, error) {
	timeout := 10 * time.Second
	client := http.Client{Timeout: timeout}
	res, err := client.Get("https://api.github.com/repos/Lyxot/CloudflareSpeedTestDNS/releases/latest")
	if err != nil {
		return "", err
	}
	// 读取资源数据 body: []byte
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	// 关闭资源流
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			utils.LogError("关闭版本检查响应流失败，错误信息: %v", err)
		}
	}(res.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if tagName, ok := result["tag_name"].(string); ok {
		return tagName, nil
	}
	return "", fmt.Errorf("can't get tag_name from github api")
}

func pgo() {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		utils.LogFatal("could not create CPU profile: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			utils.LogFatal("could not close CPU profile: %v", err)
		}
	}(f)
	if err := pprof.StartCPUProfile(f); err != nil {
		utils.LogFatal("could not start CPU profile: %v", err)
	}
	defer pprof.StopCPUProfile()
}
