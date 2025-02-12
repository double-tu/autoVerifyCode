package config

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    
    "gopkg.in/yaml.v3"
)

type Config struct {
    ServiceProvider  string `yaml:"service_provider"`
    APIToken        string `yaml:"api_token"`
    RepositoryPath  string `yaml:"repository_path"`
    PollingInterval int    `yaml:"polling_interval"`
    AutoFill       bool   `yaml:"auto_fill"`

    filePath string     // 配置文件路径
    mu       sync.Mutex // 保护并发访问
}

func LoadConfig() (*Config, error) {
    // 获取当前执行目录
    dir, err := os.Getwd()
    if err != nil {
        return nil, fmt.Errorf("获取工作目录失败: %w", err)
    }

    configPath := filepath.Join(dir, "config.yaml")
    
    // 检查配置文件是否存在
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        // 创建默认配置
        return createDefaultConfig(configPath)
    }

    // 读取配置文件
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("读取配置文件失败: %w", err)
    }

    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }

    // 验证配置
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("配置验证失败: %w", err)
    }

    config.filePath = configPath
    return &config, nil
}

func createDefaultConfig(path string) (*Config, error) {
    config := &Config{
        ServiceProvider:  "github",
        PollingInterval: 60,
        AutoFill:       true,
        filePath:       path,
    }

    // 保存默认配置
    if err := config.Save(); err != nil {
        return nil, fmt.Errorf("保存默认配置失败: %w", err)
    }

    return config, nil
}

func validateConfig(config *Config) error {
    if config.ServiceProvider != "github" && config.ServiceProvider != "gitee" {
        return errors.New("service_provider 必须是 github 或 gitee")
    }
    if config.APIToken == "" {
        return errors.New("api_token 不能为空")
    }
    if config.RepositoryPath == "" {
        return errors.New("repository_path 不能为空")
    }
    if config.PollingInterval < 1 {
        return errors.New("polling_interval 必须大于0")
    }
    return nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
    c.mu.Lock()
    defer c.mu.Unlock()

    data, err := yaml.Marshal(c)
    if err != nil {
        return fmt.Errorf("序列化配置失败: %w", err)
    }

    if err := os.WriteFile(c.filePath, data, 0644); err != nil {
        return fmt.Errorf("写入配置文件失败: %w", err)
    }

    return nil
}

// UpdateServiceProvider 更新服务提供商
func (c *Config) UpdateServiceProvider(provider string) error {
    if provider != "github" && provider != "gitee" {
        return errors.New("无效的服务提供商")
    }
    
    c.mu.Lock()
    c.ServiceProvider = provider
    c.mu.Unlock()
    
    return c.Save()
}

// UpdateAutoFill 更新自动填充设置
func (c *Config) UpdateAutoFill(enabled bool) error {
    c.mu.Lock()
    c.AutoFill = enabled
    c.mu.Unlock()
    
    return c.Save()
}

// 添加配置实时校验方法
func (c *Config) Validate() error {
    return validateConfig(c)
}

// 添加配置更新方法
func (c *Config) UpdatePollingInterval(interval int) error {
    if interval < 1 {
        return errors.New("轮询间隔必须大于0")
    }
    
    c.mu.Lock()
    c.PollingInterval = interval
    c.mu.Unlock()
    
    return c.Save()
}

func (c *Config) UpdateAPIToken(token string) error {
    if token == "" {
        return errors.New("API令牌不能为空")
    }
    
    c.mu.Lock()
    c.APIToken = token
    c.mu.Unlock()
    
    return c.Save()
}

func (c *Config) UpdateRepositoryPath(path string) error {
    if path == "" {
        return errors.New("仓库路径不能为空")
    }
    
    c.mu.Lock()
    c.RepositoryPath = path
    c.mu.Unlock()
    
    return c.Save()
}