package service

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/your-username/verify-code-tool/internal/config"
)

type VerifyCode struct {
    Date       string `json:"date"`
    VerifyCode string `json:"verifyCode"`
}

type APIService struct {
    config *config.Config
    client *http.Client
}

func NewAPIService(cfg *config.Config) *APIService {
    return &APIService{
        config: cfg,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (s *APIService) GetVerifyCode() (*VerifyCode, error) {
    var lastErr error
    for i := 0; i < 3; i++ {
        code, err := s.tryGetVerifyCode()
        if err == nil {
            return code, nil
        }
        lastErr = err
        time.Sleep(time.Duration(1<<uint(i)) * time.Second) // 指数退避
    }
    return nil, fmt.Errorf("重试3次后仍然失败: %w", lastErr)
}

func (s *APIService) tryGetVerifyCode() (*VerifyCode, error) {
    url := s.getAPIURL()
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %v", err)
    }

    // 添加认证头
    req.Header.Set("Authorization", fmt.Sprintf("token %s", s.config.APIToken))
    
    // 执行请求，包含重试逻辑
    var resp *http.Response
    for i := 0; i < 3; i++ {
        resp, err = s.client.Do(req)
        if err == nil {
            break
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    if err != nil {
        return nil, fmt.Errorf("请求失败: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API返回错误状态码: %d", resp.StatusCode)
    }

    // 读取并解析响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("读取响应失败: %v", err)
    }

    // 解析API响应
    var apiResp struct {
        Content string `json:"content"`
    }
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("解析API响应失败: %v", err)
    }

    // 解析验证码数据
    var verifyCode VerifyCode
    if err := json.Unmarshal([]byte(apiResp.Content), &verifyCode); err != nil {
        return nil, fmt.Errorf("解析验证码数据失败: %v", err)
    }

    return &verifyCode, nil
}

func (s *APIService) getAPIURL() string {
    switch s.config.ServiceProvider {
    case "github":
        return fmt.Sprintf("https://api.github.com/repos/%s/contents/code.json", s.config.RepositoryPath)
    case "gitee":
        return fmt.Sprintf("https://gitee.com/api/v5/repos/%s/contents/code.json", s.config.RepositoryPath)
    default:
        return ""
    }
} 