package ddns

import (
	"fmt"
	"strconv"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

// AliDNSConfig 阿里云DNS配置
type aliConfig struct {
	AccessKeyID     string `toml:"accesskey_id"`     // 阿里云 Access Key ID
	AccessKeySecret string `toml:"accesskey_secret"` // 阿里云 Access Key Secret
	Domain          string `toml:"domain"`           // 阿里云域名
	Subdomain       string `toml:"subdomain"`        // 子域名
	TTL             int    `toml:"ttl"`              // TTL
}

// AliDNSRecord 表示一条阿里云DNS记录
type AliDNSRecord struct {
	RecordID string
	RR       string
	Type     string
	Value    string
	TTL      int
}

// 默认配置
var (
	AliDNSConfig = aliConfig{
		AccessKeyID:     "",
		AccessKeySecret: "",
		Domain:          "",
		Subdomain:       "",
		TTL:             600,
	}
)

// newAliDNSClient 创建一个新的阿里云DNS客户端
func newAliDNSClient() (*alidns.Client, error) {
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", AliDNSConfig.AccessKeyID, AliDNSConfig.AccessKeySecret)
	return client, err
}

// SyncDNSRecords 同步阿里云DNS记录
func SyncDNSRecords(ipv4Results, ipv6Results []string) error {
	if utils.Debug {
		utils.LogDebug("开始同步阿里云DNS记录")
		if len(ipv4Results) > 0 {
			utils.LogDebug("IPv4结果: %v", ipv4Results)
		}
		if len(ipv6Results) > 0 {
			utils.LogDebug("IPv6结果: %v", ipv6Results)
		}
	}
	if AliDNSConfig.AccessKeyID == "" || AliDNSConfig.AccessKeySecret == "" || AliDNSConfig.Domain == "" || AliDNSConfig.Subdomain == "" {
		return fmt.Errorf("阿里云DNS配置不完整")
	}

	client, err := newAliDNSClient()
	if err != nil {
		return fmt.Errorf("创建阿里云DNS客户端失败: %v", err)
	}

	// 同步A记录
	if len(ipv4Results) > 0 {
		v4Records, err := getAliRecords(client, AliDNSConfig.Domain, AliDNSConfig.Subdomain, "A")
		if err != nil {
			return fmt.Errorf("获取阿里云A记录失败: %v", err)
		}
		if err := syncAliRecords(client, AliDNSConfig.Subdomain, "A", ipv4Results, v4Records); err != nil {
			return fmt.Errorf("同步阿里云A记录失败: %v", err)
		}
	}

	// 同步AAAA记录
	if len(ipv6Results) > 0 {
		v6Records, err := getAliRecords(client, AliDNSConfig.Domain, AliDNSConfig.Subdomain, "AAAA")
		if err != nil {
			return fmt.Errorf("获取阿里云AAAA记录失败: %v", err)
		}
		if err := syncAliRecords(client, AliDNSConfig.Subdomain, "AAAA", ipv6Results, v6Records); err != nil {
			return fmt.Errorf("同步阿里云AAAA记录失败: %v", err)
		}
	}

	return nil
}

// syncAliRecords 同步阿里云DNS记录
func syncAliRecords(client *alidns.Client, rr, dnsType string, desiredValues []string, existingRecords []AliDNSRecord) error {
	// 1) 跳过已存在且值一致的记录
	desiredCounter := make(map[string]int)
	for _, v := range desiredValues {
		desiredCounter[v]++
	}

	changeableRecords := []AliDNSRecord{}

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
				utils.LogDebug("更新阿里云 %s 记录: %s -> %s (ID: %s)", dnsType, rec.Value, newVal, rec.RecordID)
			}
			if err := updateAliRecord(client, rr, dnsType, newVal, rec.RecordID, AliDNSConfig.TTL); err != nil {
				return err
			}
		}
	}

	// 4) 若仍有剩余需要添加的值，则执行add
	for _, v := range remainingNeeded[updates:] {
		if utils.Debug {
			utils.LogDebug("添加阿里云 %s 记录: %s", dnsType, v)
		}
		if err := addAliRecord(client, rr, dnsType, v, AliDNSConfig.TTL); err != nil {
			return err
		}
	}

	// 5) 若仍有多余记录（未用于更新），删除之
	for _, rec := range changeableRecords[updates:] {
		if utils.Debug {
			utils.LogDebug("删除阿里云 %s 记录: %s (ID: %s)", dnsType, rec.Value, rec.RecordID)
		}
		if err := deleteAliRecord(client, rec.RecordID); err != nil {
			return err
		}
	}

	return nil
}

// getAliRecords 获取指定类型的阿里云DNS记录
func getAliRecords(client *alidns.Client, domain, rr, recordType string) ([]AliDNSRecord, error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domain
	request.RRKeyWord = rr
	request.Type = recordType

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		return nil, err
	}

	records := []AliDNSRecord{}
	for _, r := range response.DomainRecords.Record {
		if r.RR == rr && r.Type == recordType {
			records = append(records, AliDNSRecord{
				RecordID: r.RecordId,
				RR:       r.RR,
				Type:     r.Type,
				Value:    r.Value,
				TTL:      int(r.TTL),
			})
		}
	}

	return records, nil
}

// addAliRecord 添加阿里云DNS记录
func addAliRecord(client *alidns.Client, rr, recordType, value string, ttl int) error {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.DomainName = AliDNSConfig.Domain
	request.RR = rr
	request.Type = recordType
	request.Value = value
	request.TTL = requests.Integer(strconv.FormatInt(int64(ttl), 10))

	_, err := client.AddDomainRecord(request)
	return err
}

// updateAliRecord 更新阿里云DNS记录
func updateAliRecord(client *alidns.Client, rr, recordType, value, recordID string, ttl int) error {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordID
	request.RR = rr
	request.Type = recordType
	request.Value = value
	request.TTL = requests.Integer(strconv.FormatInt(int64(ttl), 10))

	_, err := client.UpdateDomainRecord(request)
	return err
}

// deleteAliRecord 删除阿里云DNS记录
func deleteAliRecord(client *alidns.Client, recordID string) error {
	request := alidns.CreateDeleteDomainRecordRequest()
	request.Scheme = "https"
	request.RecordId = recordID

	_, err := client.DeleteDomainRecord(request)
	return err
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}