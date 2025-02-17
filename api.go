package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GetGithubCode 从 GitHub 获取验证码
func GetGithubCode(token, repository, path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repository, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 记录原始响应
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %v", err)
	}
	log.Printf("GitHub API 原始响应: %s", string(rawBody))

	var result struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}

	// 重新设置响应体以便后续解码
	resp.Body = io.NopCloser(bytes.NewReader(rawBody))
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Encoding != "base64" {
		return "", fmt.Errorf("不支持的编码格式: %s", result.Encoding)
	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	var verifyCode VerifyCode
	if err := json.Unmarshal(content, &verifyCode); err != nil {
		return "", fmt.Errorf("解析验证码数据失败: %v", err)
	}

	// 验证日期是否过期
	dateStr := strings.TrimSpace(verifyCode.Date)
	if dateStr == "" {
		return "", fmt.Errorf("验证码日期为空")
	}
	
	codeTime, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return "", fmt.Errorf("解析日期失败: %v (日期字符串: %q)", err, dateStr)
	}

	if time.Now().After(codeTime) {
		return "", fmt.Errorf("验证码已过期")
	}

	return verifyCode.VerifyCode, nil
}

// GetGiteeCode 从 Gitee 获取验证码
func GetGiteeCode(token, repository, path string) (string, error) {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/contents/%s?access_token=%s", repository, path, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gitee API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 如果 encoding 字段为空，默认为 base64
	if result.Encoding == "" {
		result.Encoding = "base64"
	} else if result.Encoding != "base64" {
		return "", fmt.Errorf("不支持的编码格式: %s", result.Encoding)
	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	var verifyCode VerifyCode
	if err := json.Unmarshal(content, &verifyCode); err != nil {
		return "", fmt.Errorf("解析验证码数据失败: %v", err)
	}

	// 验证日期是否过期
	dateStr := strings.TrimSpace(verifyCode.Date)
	if dateStr == "" {
		return "", fmt.Errorf("验证码日期为空")
	}

	
	codeTime, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return "", fmt.Errorf("解析日期失败: %v (日期字符串: %q)", err, dateStr)
	}

	if time.Now().After(codeTime) {
		return "", fmt.Errorf("验证码已过期")
	}

	return verifyCode.VerifyCode, nil
}

// 只保留小写版本的函数
func getGithubCode(token, repository, path string) (VerifyCode, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repository, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return VerifyCode{}, fmt.Errorf("GitHub API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return VerifyCode{}, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Encoding != "base64" {
		return VerifyCode{}, fmt.Errorf("不支持的编码格式: %s", result.Encoding)
	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("base64解码失败: %v", err)
	}

	var verifyCode VerifyCode
	if err := json.Unmarshal(content, &verifyCode); err != nil {
		return VerifyCode{}, fmt.Errorf("解析验证码数据失败: %v", err)
	}

	return verifyCode, nil
}

func getGiteeCode(token, repository, path string) (VerifyCode, error) {
	url := fmt.Sprintf("https://gitee.com/api/v5/repos/%s/contents/%s", repository, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return VerifyCode{}, fmt.Errorf("Gitee API 返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content string `json:"content"`
		Encoding string `json:"encoding"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return VerifyCode{}, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Encoding != "base64" {
		return VerifyCode{}, fmt.Errorf("不支持的编码格式: %s", result.Encoding)
	}

	content, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return VerifyCode{}, fmt.Errorf("base64解码失败: %v", err)
	}

	var verifyCode VerifyCode
	if err := json.Unmarshal(content, &verifyCode); err != nil {
		return VerifyCode{}, fmt.Errorf("解析验证码数据失败: %v", err)
	}

	return verifyCode, nil
}