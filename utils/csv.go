package utils

import (
	"encoding/csv"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultOutput         = "result.csv"
	maxDelay              = 9999 * time.Millisecond
	minDelay              = 0 * time.Millisecond
	maxLossRate   float32 = 1.0
)

var (
	InputMaxDelay    = maxDelay
	InputMinDelay    = minDelay
	InputMaxLossRate = maxLossRate
	Output           = defaultOutput
	PrintNum         = 10
	Debug            = false // 是否开启调试模式
)

// NoPrintResult 是否打印测试结果
func NoPrintResult() bool {
	return PrintNum == 0
}

// 是否输出到文件
func noOutput() bool {
	return Output == "" || Output == " "
}

type PingData struct {
	IP          *net.IPAddr
	Transmitted int
	Received    int
	Delay       time.Duration
	Colo        string
}

type CloudflareIPData struct {
	*PingData
	lossRate      float32
	DownloadSpeed float64
}

// 计算丢包率
func (cf *CloudflareIPData) getLossRate() float32 {
	if cf.lossRate == 0 {
		pingLost := cf.Transmitted - cf.Received
		cf.lossRate = float32(pingLost) / float32(cf.Transmitted)
	}
	return cf.lossRate
}

func (cf *CloudflareIPData) toString() []string {
	result := make([]string, 7)
	result[0] = cf.IP.String()
	result[1] = strconv.Itoa(cf.Transmitted)
	result[2] = strconv.Itoa(cf.Received)
	result[3] = strconv.FormatFloat(float64(cf.getLossRate()), 'f', 2, 32)
	result[4] = strconv.FormatFloat(cf.Delay.Seconds()*1000, 'f', 2, 32)
	result[5] = strconv.FormatFloat(cf.DownloadSpeed/1024/1024, 'f', 2, 32)
	// 如果 Colo 为空，则使用 "N/A" 表示
	if cf.Colo == "" {
		result[6] = "N/A"
	} else {
		result[6] = cf.Colo
	}
	return result
}

func ExportCsv(data []CloudflareIPData) {
	if noOutput() || len(data) == 0 {
		return
	}
	fp, err := os.Create(Output)
	if err != nil {
		LogError("创建文件[%s]失败：%v", Output, err)
		return
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			LogError("关闭文件[%s]失败：%v", Output, err)
		}
	}(fp)
	w := csv.NewWriter(fp) //创建一个新的写入文件流
	_ = w.Write([]string{"IP 地址", "已发送", "已接收", "丢包率", "平均延迟", "下载速度(MB/s)", "地区码"})
	_ = w.WriteAll(convertToString(data))
	w.Flush()
}

func GetFilenameWithSuffix(filename, suffix string) string {
	if len(filename) == 0 {
		return ""
	}
	extIndex := strings.LastIndex(filename, ".")
	if extIndex == -1 {
		return filename + "_" + suffix
	}
	return filename[:extIndex] + "_" + suffix + filename[extIndex:]
}

func convertToString(data []CloudflareIPData) [][]string {
	result := make([][]string, 0)
	for _, v := range data {
		result = append(result, v.toString())
	}
	return result
}

// PingDelaySet 延迟丢包排序
type PingDelaySet []CloudflareIPData

// FilterDelay 延迟条件过滤
func (s PingDelaySet) FilterDelay() (data PingDelaySet) {
	if InputMaxDelay > maxDelay || InputMinDelay < minDelay { // 当输入的延迟条件不在默认范围内时，不进行过滤
		return s
	}
	if InputMaxDelay == maxDelay && InputMinDelay == minDelay { // 当输入的延迟条件为默认值时，不进行过滤
		return s
	}
	for _, v := range s {
		if v.Delay > InputMaxDelay { // 平均延迟上限，延迟大于条件最大值时，后面的数据都不满足条件，直接跳出循环
			break
		}
		if v.Delay < InputMinDelay { // 平均延迟下限，延迟小于条件最小值时，不满足条件，跳过
			continue
		}
		data = append(data, v) // 延迟满足条件时，添加到新数组中
	}
	return
}

// FilterLossRate 丢包条件过滤
func (s PingDelaySet) FilterLossRate() (data PingDelaySet) {
	if InputMaxLossRate >= maxLossRate { // 当输入的丢包条件为默认值时，不进行过滤
		return s
	}
	for _, v := range s {
		if v.getLossRate() > InputMaxLossRate { // 丢包几率上限
			break
		}
		data = append(data, v) // 丢包率满足条件时，添加到新数组中
	}
	return
}

func (s PingDelaySet) Len() int {
	return len(s)
}
func (s PingDelaySet) Less(i, j int) bool {
	iRate, jRate := s[i].getLossRate(), s[j].getLossRate()
	if iRate != jRate {
		return iRate < jRate
	}
	return s[i].Delay < s[j].Delay
}
func (s PingDelaySet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// DownloadSpeedSet 下载速度排序
type DownloadSpeedSet []CloudflareIPData

func (s DownloadSpeedSet) Len() int {
	return len(s)
}
func (s DownloadSpeedSet) Less(i, j int) bool {
	return s[i].DownloadSpeed > s[j].DownloadSpeed
}
func (s DownloadSpeedSet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// FilterIPv4 过滤出 IPv4 数据
func (s DownloadSpeedSet) FilterIPv4() []IPData {
	var result []IPData
	for _, data := range s {
		ip := data.IP.String()
		if strings.Contains(ip, ".") { // IPv4地址包含点
			result = append(result, IPData{
				IP:       ip,
				Packets:  data.Transmitted,
				Received: data.Received,
				LossRate: data.getLossRate(),
				Delay:    int64(data.Delay / time.Millisecond),
				Speed:    data.DownloadSpeed / 1024 / 1024, // 转为 MB/s
				Colo:     data.Colo,
			})
		}
	}
	return result
}

// FilterIPv6 过滤出 IPv6 数据
func (s DownloadSpeedSet) FilterIPv6() []IPData {
	var result []IPData
	for _, data := range s {
		ip := data.IP.String()
		if strings.Contains(ip, ":") { // IPv6地址包含冒号
			result = append(result, IPData{
				IP:       ip,
				Packets:  data.Transmitted,
				Received: data.Received,
				LossRate: data.getLossRate(),
				Delay:    int64(data.Delay / time.Millisecond),
				Speed:    data.DownloadSpeed / 1024 / 1024, // 转为 MB/s
				Colo:     data.Colo,
			})
		}
	}
	return result
}

// IPData 用于导出到 Cloudflare KV 的数据结构
type IPData struct {
	IP       string  // IP 地址
	Packets  int     // 发送的包数
	Received int     // 接收的包数
	LossRate float32 // 丢包率
	Delay    int64   // 延迟（毫秒）
	Speed    float64 // 下载速度（MB/s）
	Colo     string  // 地区码
}

func (s DownloadSpeedSet) Print() {
	if NoPrintResult() {
		return
	}
	if len(s) <= 0 { // IP数组长度(IP数量) 大于 0 时继续
		LogInfo("完整测速结果 IP 数量为 0，跳过输出结果。")
		return
	}
	dataString := convertToString(s) // 转为多维数组 [][]String
	if len(dataString) < PrintNum {  // 如果IP数组长度(IP数量) 小于  打印次数，则次数改为IP数量
		PrintNum = len(dataString)
	}
	headFormat := "%-16s%-5s%-5s%-5s%-6s%-12s%-5s"
	dataFormat := "%-18s%-8s%-8s%-8s%-10s%-16s%-8s"
	for i := 0; i < PrintNum; i++ { // 如果要输出的 IP 中包含 IPv6，那么就需要调整一下间隔
		if len(dataString[i][0]) > 15 {
			headFormat = "%-40s%-5s%-5s%-5s%-6s%-12s%-5s"
			dataFormat = "%-42s%-8s%-8s%-8s%-10s%-16s%-8s"
			break
		}
	}
	LogInfo(headFormat, "IP 地址", "已发送", "已接收", "丢包率", "平均延迟", "下载速度(MB/s)", "地区码")
	for i := 0; i < PrintNum; i++ {
		LogInfo(dataFormat, dataString[i][0], dataString[i][1], dataString[i][2], dataString[i][3], dataString[i][4], dataString[i][5], dataString[i][6])
	}
	if !noOutput() {
		LogInfo("完整测速结果已写入 %v 文件，可使用记事本/表格软件查看。", Output)
	}
}
