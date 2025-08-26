package ddns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Lyxot/CloudflareSpeedTestDNS/utils"
)

// CloudflareConfig Cloudflare DNS配置
type cloudflareConfig struct {
	APIToken  string `toml:"api_token"`  // Cloudflare API Token
	ZoneID    string `toml:"zone_id"`    // Cloudflare Zone ID
	Domain    string `toml:"domain"`     // 域名
	Subdomain string `toml:"subdomain"` // 子域名
	Proxied   bool   `toml:"proxied"`    // 是否开启Cloudflare代理
	TTL       int    `toml:"ttl"`        // TTL，1为自动
}

// CloudflareDNSRecord 表示一条Cloudflare DNS记录
type CloudflareDNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

// 默认配置
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

// Cloudflare API地址
const (
	CloudflareAPI = "https://api.cloudflare.com/client/v4"
)

// SyncCloudflareRecords 同步Cloudflare DNS记录
func SyncCloudflareRecords(ipv4Results, ipv6Results []string) error {
	if utils.Debug {
		utils.Yellow.Printf("[调试] 开始同步Cloudflare DNS记录\n")
		if len(ipv4Results) > 0 {
			utils.Yellow.Printf("[调试] IPv4结果: %v\n", ipv4Results)
		}
		if len(ipv6Results) > 0 {
			utils.Yellow.Printf("[调试] IPv6结果: %v\n", ipv6Results)
		}
	}
	if CloudflareConfig.APIToken == "" || CloudflareConfig.ZoneID == "" || CloudflareConfig.Domain == "" {
		return fmt.Errorf("Cloudflare DNS配置不完整")
	}

	// 同步A记录
	if len(ipv4Results) > 0 {
		v4Records, err := getCloudflareRecords("A")
		if err != nil {
			return fmt.Errorf("获取Cloudflare A记录失败: %v", err)
		}
		if err := syncCloudflareRecords("A", ipv4Results, v4Records); err != nil {
			return fmt.Errorf("同步Cloudflare A记录失败: %v", err)
		}
	}

	// 同步AAAA记录
	if len(ipv6Results) > 0 {
		v6Records, err := getCloudflareRecords("AAAA")
		if err != nil {
			return fmt.Errorf("获取Cloudflare AAAA记录失败: %v", err)
		}
		if err := syncCloudflareRecords("AAAA", ipv6Results, v6Records); err != nil {
			return fmt.Errorf("同步Cloudflare AAAA记录失败: %v", err)
		}
	}

	return nil
}

// syncCloudflareRecords 同步Cloudflare DNS记录
func syncCloudflareRecords(recordType string, desiredValues []string, existingRecords []CloudflareDNSRecord) error {
	// 1) 跳过已存在且值一致的记录
	desiredCounter := make(map[string]int)
	for _, v := range desiredValues {
		desiredCounter[v]++
	}

	changeableRecords := []CloudflareDNSRecord{}

	for _, rec := range existingRecords {
		if count, exists := desiredCounter[rec.Content]; exists && count > 0 {
			desiredCounter[rec.Content]--
			// 记录已经存在，无需操作
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
		if rec.Content != newVal {
			if utils.Debug {
				utils.Yellow.Printf("[调试] 更新Cloudflare %s记录: %s -> %s (ID: %s)\n", recordType, rec.Content, newVal, rec.ID)
			}
			if err := updateCloudflareRecord(recordType, newVal, rec.ID); err != nil {
				return err
			}
		}
	}

	// 4) 若仍有剩余需要添加的值，则执行add
	for _, v := range remainingNeeded[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 添加Cloudflare %s记录: %s\n", recordType, v)
		}
		if err := addCloudflareRecord(recordType, v); err != nil {
			return err
		}
	}

	// 5) 若仍有多余记录（未用于更新），删除之
	for _, rec := range changeableRecords[updates:] {
		if utils.Debug {
			utils.Yellow.Printf("[调试] 删除Cloudflare %s记录: %s (ID: %s)\n", recordType, rec.Content, rec.ID)
		}
		if err := deleteCloudflareRecord(rec.ID); err != nil {
			return err
		}
	}

	return nil
}

// getCloudflareRecords 获取指定类型的Cloudflare DNS记录
func getCloudflareRecords(recordType string) ([]CloudflareDNSRecord, error) {
	// 根据子域名和域名拼接得到完整域名
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		// 如果子域名为空或@，则使用域名本身
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	url := fmt.Sprintf("%s/zones/%s/dns_records?type=%s&name=%s", 
		CloudflareAPI, CloudflareConfig.ZoneID, recordType, fullName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+CloudflareConfig.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if utils.Debug {
		utils.Yellow.Printf("[调试] Cloudflare API响应: %s\n", string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result []CloudflareDNSRecord `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("Cloudflare API错误: %s (代码: %d)",
				result.Errors[0].Message, result.Errors[0].Code)
		}
		return nil, fmt.Errorf("Cloudflare API请求失败")
	}

	return result.Result, nil
}

// addCloudflareRecord 添加Cloudflare DNS记录
func addCloudflareRecord(recordType, content string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records", CloudflareAPI, CloudflareConfig.ZoneID)

	// 根据子域名和域名拼接得到完整域名
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		// 如果子域名为空或@，则使用域名本身
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	data := map[string]interface{}{
		"type":    recordType,
		"name":    fullName,
		"content": content,
		"ttl":     CloudflareConfig.TTL,
		"proxied": CloudflareConfig.Proxied,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+CloudflareConfig.APIToken)
	req.Header.Set("Content-Type", "application/json")

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
		utils.Yellow.Printf("[调试] Cloudflare API添加记录响应: %s\n", string(body))
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
			return fmt.Errorf("Cloudflare API错误: %s (代码: %d)",
				result.Errors[0].Message, result.Errors[0].Code)
		}
		return fmt.Errorf("Cloudflare API请求失败")
	}

	return nil
}

// updateCloudflareRecord 更新Cloudflare DNS记录
func updateCloudflareRecord(recordType, content, id string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s",
		CloudflareAPI, CloudflareConfig.ZoneID, id)

	// 根据子域名和域名拼接得到完整域名
	var fullName string
	if CloudflareConfig.Subdomain == "" || CloudflareConfig.Subdomain == "@" {
		// 如果子域名为空或@，则使用域名本身
		fullName = CloudflareConfig.Domain
	} else {
		fullName = CloudflareConfig.Subdomain + "." + CloudflareConfig.Domain
	}

	data := map[string]interface{}{
		"type":    recordType,
		"name":    fullName,
		"content": content,
		"ttl":     CloudflareConfig.TTL,
		"proxied": CloudflareConfig.Proxied,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+CloudflareConfig.APIToken)
	req.Header.Set("Content-Type", "application/json")

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
		utils.Yellow.Printf("[调试] Cloudflare API更新记录响应: %s\n", string(body))
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
			return fmt.Errorf("Cloudflare API错误: %s (代码: %d)",
				result.Errors[0].Message, result.Errors[0].Code)
		}
		return fmt.Errorf("Cloudflare API请求失败")
	}

	return nil
}

// deleteCloudflareRecord 删除Cloudflare DNS记录
func deleteCloudflareRecord(id string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s",
		CloudflareAPI, CloudflareConfig.ZoneID, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+CloudflareConfig.APIToken)
	req.Header.Set("Content-Type", "application/json")

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
		utils.Yellow.Printf("[调试] Cloudflare API删除记录响应: %s\n", string(body))
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
			return fmt.Errorf("Cloudflare API错误: %s (代码: %d)",
				result.Errors[0].Message, result.Errors[0].Code)
		}
		return fmt.Errorf("Cloudflare API请求失败")
	}

	return nil
}
