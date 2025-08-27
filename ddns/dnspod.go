
package ddns

import (
	"fmt"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// DNSPodConfig DNSPod DNS配置
type dnspodConfig struct {
	SecretID  string `toml:"secret_id"`  // DNSPod Secret ID
	SecretKey string `toml:"secret_key"` // DNSPod Secret Key
	Domain    string `toml:"domain"`     // 域名
	Subdomain string `toml:"subdomain"`  // 子域名
	TTL       int    `toml:"ttl"`        // TTL
}

// DNSPodRecord 表示一条DNSPod DNS记录
type DNSPodRecord struct {
	ID    uint64
	Name  string
	Type  string
	Value string
	TTL   uint64
}

// 默认配置
var (
	DNSPodConfig = dnspodConfig{
		SecretID:  "",
		SecretKey: "",
		Domain:    "",
		Subdomain: "",
		TTL:       600,
	}
)

// newDNSPodClient 创建一个新的DNSPod客户端
func newDNSPodClient() (*dnspod.Client, error) {
	credential := common.NewCredential(
		DNSPodConfig.SecretID,
		DNSPodConfig.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	client, err := dnspod.NewClient(credential, "", cpf)
	return client, err
}

// SyncDNSPodRecords 同步DNSPod DNS记录
func SyncDNSPodRecords(ipv4Results, ipv6Results []string) error {
	if utils.Debug {
		utils.Yellow.Printf("[调试] 开始同步DNSPod DNS记录\n")
		if len(ipv4Results) > 0 {
			utils.Yellow.Printf("[调试] IPv4结果: %v\n", ipv4Results)
		}
		if len(ipv6Results) > 0 {
			utils.Yellow.Printf("[调试] IPv6结果: %v\n", ipv6Results)
		}
	}
	if DNSPodConfig.SecretID == "" || DNSPodConfig.SecretKey == "" || DNSPodConfig.Domain == "" || DNSPodConfig.Subdomain == "" {
		return fmt.Errorf("DNSPod DNS配置不完整")
	}

	client, err := newDNSPodClient()
	if err != nil {
		return fmt.Errorf("创建DNSPod客户端失败: %v", err)
	}

	// 同步A记录
	if len(ipv4Results) > 0 {
		v4Records, err := getDNSPodRecords(client, DNSPodConfig.Domain, DNSPodConfig.Subdomain, "A")
		if err != nil {
			return fmt.Errorf("获取DNSPod A记录失败: %v", err)
		}
		if err := syncDNSPodRecords(client, DNSPodConfig.Subdomain, "A", ipv4Results, v4Records); err != nil {
			return fmt.Errorf("同步DNSPod A记录失败: %v", err)
		}
	}

	// 同步AAAA记录
	if len(ipv6Results) > 0 {
		v6Records, err := getDNSPodRecords(client, DNSPodConfig.Domain, DNSPodConfig.Subdomain, "AAAA")
		if err != nil {
			return fmt.Errorf("获取DNSPod AAAA记录失败: %v", err)
		}
		if err := syncDNSPodRecords(client, DNSPodConfig.Subdomain, "AAAA", ipv6Results, v6Records); err != nil {
			return fmt.Errorf("同步DNSPod AAAA记录失败: %v", err)
		}
	}

	return nil
}

// syncDNSPodRecords 同步DNSPod DNS记录
func syncDNSPodRecords(client *dnspod.Client, rr, dnsType string, desiredValues []string, existingRecords []DNSPodRecord) error {
	// 1) 跳过已存在且值一致的记录
	desiredCounter := make(map[string]int)
	for _, v := range desiredValues {
		desiredCounter[v]++
	}

	changeableRecords := []DNSPodRecord{}

	for _, rec := range existingRecords {
		if count, exists := desiredCounter[rec.Value]; exists && count > 0 {
			desiredCounter[rec.Value]--
		} else {
			changeableRecords = append(changeableRecords, rec)
		}
	}

	// 2) 展开剩余需要的目标值
	remainingNeeded := []string{}
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
		if rec.Value != newVal {
			if utils.Debug {
				utils.Yellow.Printf("[调试] 更新DNSPod %s 记录: %s -> %s (ID: %d)\n", dnsType, rec.Value, newVal, rec.ID)
			}
			if err := updateDNSPodRecord(client, rec.ID, rr, dnsType, newVal, uint64(DNSPodConfig.TTL)); err != nil {
				return err
			}
		}
	}

	// 4) 若仍有剩余需要添加的值，则执行add
	for _, v := range remainingNeeded[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 添加DNSPod %s 记录: %s\n", dnsType, v)
		}
		if err := addDNSPodRecord(client, rr, dnsType, v, uint64(DNSPodConfig.TTL)); err != nil {
			return err
		}
	}

	// 5) 若仍有多余记录（未用于更新），删除之
	for _, rec := range changeableRecords[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 删除DNSPod %s 记录: %s (ID: %d)\n", dnsType, rec.Value, rec.ID)
		}
		if err := deleteDNSPodRecord(client, rec.ID); err != nil {
			return err
		}
	}

	return nil
}

// getDNSPodRecords 获取指定类型的DNSPod DNS记录
func getDNSPodRecords(client *dnspod.Client, domain, rr, recordType string) ([]DNSPodRecord, error) {
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)
	request.Subdomain = common.StringPtr(rr)
	request.RecordType = common.StringPtr(recordType)

	response, err := client.DescribeRecordList(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			if sdkErr.Code == "ResourceNotFound.NoDataOfRecord" {
				return []DNSPodRecord{}, nil // No records found, but not an error
			}
		}
		return nil, fmt.Errorf("API error: %v", err)
	}

	records := []DNSPodRecord{}
	for _, r := range response.Response.RecordList {
		records = append(records, DNSPodRecord{
			ID:    *r.RecordId,
			Name:  *r.Name,
			Type:  *r.Type,
			Value: *r.Value,
			TTL:   *r.TTL,
		})
	}

	return records, nil
}

// addDNSPodRecord 添加DNSPod DNS记录
func addDNSPodRecord(client *dnspod.Client, rr, recordType, value string, ttl uint64) error {
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(DNSPodConfig.Domain)
	request.SubDomain = common.StringPtr(rr)
	request.RecordType = common.StringPtr(recordType)
	request.Value = common.StringPtr(value)
	request.RecordLine = common.StringPtr("默认")
	request.TTL = common.Uint64Ptr(ttl)

	_, err := client.CreateRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API error: %v", err)
	}
	return err
}

// updateDNSPodRecord 更新DNSPod DNS记录
func updateDNSPodRecord(client *dnspod.Client, recordID uint64, rr, recordType, value string, ttl uint64) error {
	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(DNSPodConfig.Domain)
	request.RecordId = common.Uint64Ptr(recordID)
	request.SubDomain = common.StringPtr(rr)
	request.RecordType = common.StringPtr(recordType)
	request.Value = common.StringPtr(value)
	request.RecordLine = common.StringPtr("默认")
	request.TTL = common.Uint64Ptr(ttl)

	_, err := client.ModifyRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API error: %v", err)
	}
	return err
}

// deleteDNSPodRecord 删除DNSPod DNS记录
func deleteDNSPodRecord(client *dnspod.Client, recordID uint64) error {
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(DNSPodConfig.Domain)
	request.RecordId = common.Uint64Ptr(recordID)

	_, err := client.DeleteRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API error: %v", err)
	}
	return err
}
