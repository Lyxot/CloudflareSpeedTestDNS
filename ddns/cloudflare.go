package ddns

import (
	"context"
	"fmt"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
	"github.com/cloudflare/cloudflare-go"
)

// CloudflareConfig Cloudflare DNS配置
type cloudflareConfig struct {
	APIToken  string `toml:"api_token"` // Cloudflare API Token
	ZoneID    string `toml:"zone_id"`   // Cloudflare Zone ID
	Domain    string `toml:"domain"`    // 域名
	Subdomain string `toml:"subdomain"` // 子域名
	Proxied   bool   `toml:"proxied"`   // 是否开启Cloudflare代理
	TTL       int    `toml:"ttl"`       // TTL，1为自动
}

// CloudflareConfig 默认配置
var (
	CloudflareConfig = cloudflareConfig{
		APIToken:  "",
		ZoneID:    "",
		Domain:    "",
		Subdomain: "",
		Proxied:   false,
		TTL:       1,
	}
)

// newCloudflareClient 创建一个新的Cloudflare客户端
func newCloudflareClient() (*cloudflare.API, error) {
	api, err := cloudflare.NewWithAPIToken(CloudflareConfig.APIToken)
	return api, err
}

// SyncCloudflareRecords 同步Cloudflare DNS记录
func SyncCloudflareRecords(ipv4Results, ipv6Results []string) error {
	if utils.Debug {
		utils.LogDebug("开始同步Cloudflare DNS记录")
		if len(ipv4Results) > 0 {
			utils.LogDebug("IPv4结果: %v", ipv4Results)
		}
		if len(ipv6Results) > 0 {
			utils.LogDebug("IPv6结果: %v", ipv6Results)
		}
	}
	if CloudflareConfig.APIToken == "" || CloudflareConfig.ZoneID == "" || CloudflareConfig.Domain == "" {
		return fmt.Errorf("cloudflare DNS配置不完整")
	}

	api, err := newCloudflareClient()
	if err != nil {
		return fmt.Errorf("创建Cloudflare客户端失败: %v", err)
	}

	ctx := context.Background()

	// 同步A记录
	if len(ipv4Results) > 0 {
		v4Records, err := getCloudflareRecords(ctx, api, "A")
		if err != nil {
			return fmt.Errorf("获取Cloudflare A记录失败: %v", err)
		}
		if err := syncCloudflareRecords(ctx, api, "A", ipv4Results, v4Records); err != nil {
			return fmt.Errorf("同步Cloudflare A记录失败: %v", err)
		}
	}

	// 同步AAAA记录
	if len(ipv6Results) > 0 {
		v6Records, err := getCloudflareRecords(ctx, api, "AAAA")
		if err != nil {
			return fmt.Errorf("获取Cloudflare AAAA记录失败: %v", err)
		}
		if err := syncCloudflareRecords(ctx, api, "AAAA", ipv6Results, v6Records); err != nil {
			return fmt.Errorf("同步Cloudflare AAAA记录失败: %v", err)
		}
	}

	return nil
}

// syncCloudflareRecords 同步Cloudflare DNS记录
func syncCloudflareRecords(ctx context.Context, api *cloudflare.API, recordType string, desiredValues []string, existingRecords []cloudflare.DNSRecord) error {
	// 1) 跳过已存在且值一致的记录
	desiredCounter := make(map[string]int)
	for _, v := range desiredValues {
		desiredCounter[v]++
	}

	var changeableRecords []cloudflare.DNSRecord

	for _, rec := range existingRecords {
		if count, exists := desiredCounter[rec.Content]; exists && count > 0 {
			desiredCounter[rec.Content]--
		} else {
			changeableRecords = append(changeableRecords, rec)
		}
	}

	// 2) 展开剩余需要的目标值
	var remainingNeeded []string
	for v, c := range desiredCounter {
		for i := 0; i < c; i++ {
			remainingNeeded = append(remainingNeeded, v)
		}
	}

	// 3) 先用"多余/需要变更"的记录进行update
	updates := min(len(changeableRecords), len(remainingNeeded))
	for i := 0; i < updates; i++ {
		rec := changeableRecords[i]
		newVal := remainingNeeded[i]
		if rec.Content != newVal {
			if utils.Debug {
				utils.LogDebug("更新Cloudflare %s记录: %s -> %s (ID: %s)", recordType, rec.Content, newVal, rec.ID)
			}
			if err := updateCloudflareRecord(ctx, api, recordType, newVal, rec.ID); err != nil {
				return err
			}
		}
	}

	// 4) 若仍有剩余需要添加的值，则执行add
	for _, v := range remainingNeeded[updates:] {
		if utils.Debug {
			utils.LogDebug("添加Cloudflare %s记录: %s", recordType, v)
		}
		if err := addCloudflareRecord(ctx, api, recordType, v); err != nil {
			return err
		}
	}

	// 5) 若仍有多余记录（未用于更新），删除之
	for _, rec := range changeableRecords[updates:] {
		if utils.Debug {
			utils.LogDebug("删除Cloudflare %s记录: %s (ID: %s)", recordType, rec.Content, rec.ID)
		}
		if err := deleteCloudflareRecord(ctx, api, rec.ID); err != nil {
			return err
		}
	}

	return nil
}

// getCloudflareRecords 获取指定类型的Cloudflare DNS记录
func getCloudflareRecords(ctx context.Context, api *cloudflare.API, recordType string) ([]cloudflare.DNSRecord, error) {
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	records, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(CloudflareConfig.ZoneID), cloudflare.ListDNSRecordsParams{
		Type: recordType,
		Name: fullName,
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

// addCloudflareRecord 添加Cloudflare DNS记录
func addCloudflareRecord(ctx context.Context, api *cloudflare.API, recordType, content string) error {
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	_, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(CloudflareConfig.ZoneID), cloudflare.CreateDNSRecordParams{
		Type:    recordType,
		Name:    fullName,
		Content: content,
		TTL:     CloudflareConfig.TTL,
		Proxied: &CloudflareConfig.Proxied,
	})
	return err
}

// updateCloudflareRecord 更新Cloudflare DNS记录
func updateCloudflareRecord(ctx context.Context, api *cloudflare.API, recordType, content, id string) error {
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	_, err := api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(CloudflareConfig.ZoneID), cloudflare.UpdateDNSRecordParams{
		ID:      id,
		Type:    recordType,
		Name:    fullName,
		Content: content,
		TTL:     CloudflareConfig.TTL,
		Proxied: &CloudflareConfig.Proxied,
	})
	return err
}

// deleteCloudflareRecord 删除Cloudflare DNS记录
func deleteCloudflareRecord(ctx context.Context, api *cloudflare.API, id string) error {
	return api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(CloudflareConfig.ZoneID), id)
}
