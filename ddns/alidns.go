package ddns

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
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

// 阿里云DNS API地址
const (
	AliDNSAPI = "https://alidns.aliyuncs.com"
)

// SyncDNSRecords 同步阿里云DNS记录
func SyncDNSRecords(ipv4Results, ipv6Results []string) error {
	if utils.Debug {
		utils.Yellow.Printf("[调试] 开始同步阿里云DNS记录\n")
		if len(ipv4Results) > 0 {
			utils.Yellow.Printf("[调试] IPv4结果: %v\n", ipv4Results)
		}
		if len(ipv6Results) > 0 {
			utils.Yellow.Printf("[调试] IPv6结果: %v\n", ipv6Results)
		}
	}
	if AliDNSConfig.AccessKeyID == "" || AliDNSConfig.AccessKeySecret == "" || AliDNSConfig.Domain == "" || AliDNSConfig.Subdomain == "" {
		return fmt.Errorf("阿里云DNS配置不完整")
	}

	// 同步A记录
	if len(ipv4Results) > 0 {
		v4Records, err := getAliRecords(AliDNSConfig.Domain, AliDNSConfig.Subdomain, "A")
		if err != nil {
			return fmt.Errorf("获取阿里云A记录失败: %v", err)
		}
		if err := syncAliRecords(AliDNSConfig.Subdomain, "A", ipv4Results, v4Records); err != nil {
			return fmt.Errorf("同步阿里云A记录失败: %v", err)
		}
	}

	// 同步AAAA记录
	if len(ipv6Results) > 0 {
		v6Records, err := getAliRecords(AliDNSConfig.Domain, AliDNSConfig.Subdomain, "AAAA")
		if err != nil {
			return fmt.Errorf("获取阿里云AAAA记录失败: %v", err)
		}
		if err := syncAliRecords(AliDNSConfig.Subdomain, "AAAA", ipv6Results, v6Records); err != nil {
			return fmt.Errorf("同步阿里云AAAA记录失败: %v", err)
		}
	}

	return nil
}

// syncAliRecords 同步阿里云DNS记录
func syncAliRecords(rr, dnsType string, desiredValues []string, existingRecords []AliDNSRecord) error {
	// 1) 跳过已存在且值一致的记录
	desiredCounter := make(map[string]int)
	for _, v := range desiredValues {
		desiredCounter[v]++
	}

	// 不需要记录保留的ID
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
				utils.Yellow.Printf("[调试] 更新阿里云 %s 记录: %s -> %s (ID: %s)\n", dnsType, rec.Value, newVal, rec.RecordID)
			}
			if err := updateAliRecord(rr, dnsType, newVal, rec.RecordID, AliDNSConfig.TTL); err != nil {
				return err
			}
		}
	}

	// 4) 若仍有剩余需要添加的值，则执行add
	for _, v := range remainingNeeded[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 添加阿里云 %s 记录: %s\n", dnsType, v)
		}
		if err := addAliRecord(rr, dnsType, v, AliDNSConfig.TTL); err != nil {
			return err
		}
	}

	// 5) 若仍有多余记录（未用于更新），删除之
	for _, rec := range changeableRecords[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 删除阿里云 %s 记录: %s (ID: %s)\n", dnsType, rec.Value, rec.RecordID)
		}
		if err := deleteAliRecord(rec.RecordID); err != nil {
			return err
		}
	}

	return nil
}

// getAliRecords 获取指定类型的阿里云DNS记录
func getAliRecords(domain, rr, recordType string) ([]AliDNSRecord, error) {
	params := map[string]string{
		"Action":     "DescribeDomainRecords",
		"DomainName": domain,
	}

	resp, err := doAliRequest(params)
	if err != nil {
		return nil, err
	}

	var result struct {
		DomainRecords struct {
			Record []struct {
				RecordID string `json:"RecordId"`
				RR       string `json:"RR"`
				Type     string `json:"Type"`
				Value    string `json:"Value"`
				TTL      int    `json:"TTL"`
			} `json:"Record"`
		} `json:"DomainRecords"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, err
	}

	records := []AliDNSRecord{}
	for _, r := range result.DomainRecords.Record {
		if r.RR == rr && r.Type == recordType {
			records = append(records, AliDNSRecord{
				RecordID: r.RecordID,
				RR:       r.RR,
				Type:     r.Type,
				Value:    r.Value,
				TTL:      r.TTL,
			})
		}
	}

	return records, nil
}

// addAliRecord 添加阿里云DNS记录
func addAliRecord(rr, recordType, value string, ttl int) error {
	params := map[string]string{
		"Action":     "AddDomainRecord",
		"DomainName": AliDNSConfig.Domain,
		"RR":         rr,
		"Type":       recordType,
		"Value":      value,
		"TTL":        strconv.Itoa(ttl),
	}

	_, err := doAliRequest(params)
	return err
}

// updateAliRecord 更新阿里云DNS记录
func updateAliRecord(rr, recordType, value, recordID string, ttl int) error {
	params := map[string]string{
		"Action":   "UpdateDomainRecord",
		"RecordId": recordID,
		"RR":       rr,
		"Type":     recordType,
		"Value":    value,
		"TTL":      strconv.Itoa(ttl),
	}

	_, err := doAliRequest(params)
	return err
}

// deleteAliRecord 删除阿里云DNS记录
func deleteAliRecord(recordID string) error {
	params := map[string]string{
		"Action":   "DeleteDomainRecord",
		"RecordId": recordID,
	}

	_, err := doAliRequest(params)
	return err
}

// doAliRequest 执行阿里云API请求
func doAliRequest(params map[string]string) (string, error) {
	if utils.Debug {
		utils.Yellow.Printf("[调试] 请求阿里云DNS API: %s\n", params["Action"])
	}
	// 公共参数
	params["Format"] = "JSON"
	params["Version"] = "2015-01-09"
	params["AccessKeyId"] = AliDNSConfig.AccessKeyID
	params["SignatureMethod"] = "HMAC-SHA1"
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)

	// 计算签名
	signature := calculateAliSignature(params, AliDNSConfig.AccessKeySecret)
	params["Signature"] = signature

	// 构建请求URL
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	reqURL := AliDNSAPI + "?" + query.Encode()

	// 发送请求
	resp, err := http.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if utils.Debug {
		utils.Yellow.Printf("[调试] 阿里云DNS API响应: %s\n", string(body))
	}

	// 检查错误
	var errorResp struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Code != "" {
		return "", fmt.Errorf("%s: %s", errorResp.Code, errorResp.Message)
	}

	return string(body), nil
}

// calculateAliSignature 计算阿里云API签名
func calculateAliSignature(params map[string]string, secret string) string {
	// 按照参数名称的字典顺序排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建规范化请求字符串
	var canonicalizedQueryString strings.Builder
	for _, k := range keys {
		canonicalizedQueryString.WriteString("&")
		canonicalizedQueryString.WriteString(aliPercentEncode(k))
		canonicalizedQueryString.WriteString("=")
		canonicalizedQueryString.WriteString(aliPercentEncode(params[k]))
	}

	// 构建待签名字符串
	stringToSign := "GET&%2F&" + aliPercentEncode(canonicalizedQueryString.String()[1:])

	// 计算HMAC-SHA1签名
	key := []byte(secret + "&")
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

// aliPercentEncode URL编码
func aliPercentEncode(s string) string {
	// 实现URL编码
	// 这里简化处理，实际实现时需要按照阿里云的要求进行编码
	return url.QueryEscape(s)
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
