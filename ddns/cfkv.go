package ddns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
	"github.com/cloudflare/cloudflare-go"
)

// cloudflareKVConfig Cloudflare KV配置
type cloudflareKVConfig struct {
	Enable      bool   // 是否启用Cloudflare KV
	APIToken    string // Cloudflare API Token
	AccountID   string // Cloudflare Account ID
	NamespaceID string // Cloudflare KV Namespace ID
}

// 默认配置
var (
	CloudflareKVConfig = cloudflareKVConfig{
		Enable:      false,
		APIToken:    "",
		AccountID:   "",
		NamespaceID: "",
	}
)

// newCloudflareKVClient 创建一个新的Cloudflare客户端
func newCloudflareKVClient() (*cloudflare.API, error) {
	api, err := cloudflare.NewWithAPIToken(CloudflareKVConfig.APIToken)
	return api, err
}

// SyncCloudflareKV 同步测速结果到Cloudflare KV
func SyncCloudflareKV(ipv4Data, ipv6Data []utils.IPData) error {
	if utils.Debug {
		utils.LogDebug("开始同步数据到Cloudflare KV")
	}

	if CloudflareKVConfig.APIToken == "" || CloudflareKVConfig.AccountID == "" || CloudflareKVConfig.NamespaceID == "" {
		return fmt.Errorf("cloudflare kv配置不完整")
	}

	api, err := newCloudflareKVClient()
	if err != nil {
		return fmt.Errorf("创建Cloudflare客户端失败: %v", err)
	}

	ctx := context.Background()

	// 当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 同步IPv4数据
	if len(ipv4Data) > 0 {
		// 构建IPv4数据字符串
		ipv4DataStr := buildKVDataString(ipv4Data)

		// 写入IPv4数据
		if err := writeToCloudflareKV(ctx, api, "ipv4", ipv4DataStr); err != nil {
			return fmt.Errorf("写入IPv4数据到Cloudflare KV失败: %v", err)
		}

		// 写入IPv4更新时间
		if err := writeToCloudflareKV(ctx, api, "ipv4time", currentTime); err != nil {
			return fmt.Errorf("写入IPv4更新时间到Cloudflare KV失败: %v", err)
		}

		if utils.Debug {
			utils.LogDebug("IPv4数据已同步到Cloudflare KV")
		}
	}

	// 同步IPv6数据
	if len(ipv6Data) > 0 {
		// 构建IPv6数据字符串
		ipv6DataStr := buildKVDataString(ipv6Data)

		// 写入IPv6数据
		if err := writeToCloudflareKV(ctx, api, "ipv6", ipv6DataStr); err != nil {
			return fmt.Errorf("写入IPv6数据到Cloudflare KV失败: %v", err)
		}

		// 写入IPv6更新时间
		if err := writeToCloudflareKV(ctx, api, "ipv6time", currentTime); err != nil {
			return fmt.Errorf("写入IPv6更新时间到Cloudflare KV失败: %v", err)
		}

		if utils.Debug {
			utils.LogDebug("IPv6数据已同步到Cloudflare KV")
		}
	}

	return nil
}

// buildKVDataString 构建KV数据字符串
// 格式: {IP地址},{已发送ping包},{已接收ping包},{丢包率},{平均延迟},{下载速度},{地区码}&{下一条数据}&...
func buildKVDataString(ipData []utils.IPData) string {
	var dataItems []string

	for _, data := range ipData {
		// 构建数据项: IP,发包数,收包数,丢包率,平均延迟,下载速度,地区码
		colo := "N/A"
		if data.Colo != "" {
			colo = data.Colo
		}
		item := fmt.Sprintf("%s,%d,%d,%.2f,%d,%.2f,%s",
			data.IP,
			data.Packets,
			data.Received,
			data.LossRate*100, // 转换为百分比
			data.Delay,        // 毫秒
			data.Speed,        // MB/s
			colo)              // 如果地区码为空，则设置为"N/A"
		dataItems = append(dataItems, item)
	}

	// 用&连接所有数据项
	return strings.Join(dataItems, "&")
}

// writeToCloudflareKV 写入数据到Cloudflare KV
func writeToCloudflareKV(ctx context.Context, api *cloudflare.API, key, value string) error {
	_, err := api.WriteWorkersKVEntry(ctx, cloudflare.AccountIdentifier(CloudflareKVConfig.AccountID), cloudflare.WriteWorkersKVEntryParams{
		NamespaceID: CloudflareKVConfig.NamespaceID,
		Key:         key,
		Value:       []byte(value),
	})
	return err
}
