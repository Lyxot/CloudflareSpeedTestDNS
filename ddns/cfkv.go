package ddns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
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

// Cloudflare KV API地址
const (
	CloudflareKVAPI = "https://api.cloudflare.com/client/v4/accounts/%s/storage/kv/namespaces/%s/values"
)

// SyncCloudflareKV 同步测速结果到Cloudflare KV
func SyncCloudflareKV(ipv4Data, ipv6Data []utils.IPData) error {
	if utils.Debug {
		utils.Yellow.Printf("[调试] 开始同步数据到Cloudflare KV\n")
	}

	if CloudflareKVConfig.APIToken == "" || CloudflareKVConfig.AccountID == "" || CloudflareKVConfig.NamespaceID == "" {
		return fmt.Errorf("cloudflare kv配置不完整")
	}

	// 当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 同步IPv4数据
	if len(ipv4Data) > 0 {
		// 构建IPv4数据字符串
		ipv4DataStr := buildKVDataString(ipv4Data)

		// 写入IPv4数据
		if err := writeToCloudflareKV("ipv4", ipv4DataStr); err != nil {
			return fmt.Errorf("写入IPv4数据到Cloudflare KV失败: %v", err)
		}

		// 写入IPv4更新时间
		if err := writeToCloudflareKV("ipv4time", currentTime); err != nil {
			return fmt.Errorf("写入IPv4更新时间到Cloudflare KV失败: %v", err)
		}

		if utils.Debug {
			utils.Yellow.Printf("[调试] IPv4数据已同步到Cloudflare KV\n")
		}
	}

	// 同步IPv6数据
	if len(ipv6Data) > 0 {
		// 构建IPv6数据字符串
		ipv6DataStr := buildKVDataString(ipv6Data)

		// 写入IPv6数据
		if err := writeToCloudflareKV("ipv6", ipv6DataStr); err != nil {
			return fmt.Errorf("写入IPv6数据到Cloudflare KV失败: %v", err)
		}

		// 写入IPv6更新时间
		if err := writeToCloudflareKV("ipv6time", currentTime); err != nil {
			return fmt.Errorf("写入IPv6更新时间到Cloudflare KV失败: %v", err)
		}

		if utils.Debug {
			utils.Yellow.Printf("[调试] IPv6数据已同步到Cloudflare KV\n")
		}
	}

	return nil
}

// buildKVDataString 构建KV数据字符串
// 格式: {IP地址},{已发送ping包},{已接收ping包},{丢包率},{平均延迟},{下载速度}&{下一条数据}&...
func buildKVDataString(ipData []utils.IPData) string {
	var dataItems []string

	for _, data := range ipData {
		// 构建数据项: IP,发包数,收包数,丢包率,平均延迟,下载速度
		item := fmt.Sprintf("%s,%d,%d,%.2f,%d,%.2f",
			data.IP,
			data.Packets,
			data.Received,
			data.LossRate*100, // 转换为百分比
			data.Delay,        // 毫秒
			data.Speed)        // MB/s
		dataItems = append(dataItems, item)
	}

	// 用&连接所有数据项
	return strings.Join(dataItems, "&")
}

// writeToCloudflareKV 写入数据到Cloudflare KV
func writeToCloudflareKV(key, value string) error {
	url := fmt.Sprintf(CloudflareKVAPI, CloudflareKVConfig.AccountID, CloudflareKVConfig.NamespaceID) + "/" + key

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(value))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+CloudflareKVConfig.APIToken)
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if utils.Debug {
		utils.Yellow.Printf("[调试] Cloudflare KV API响应: %s\n", string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			return fmt.Errorf("cloudflare kv api错误: %s (代码: %d)",
				result.Errors[0].Message, result.Errors[0].Code)
		}
		return fmt.Errorf("cloudflare kv api请求失败")
	}

	return nil
}
