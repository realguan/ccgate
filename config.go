package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// Platform 表示平台配置
type Platform struct {
	Name                string `json:"name"`
	Vendor              string `json:"vendor"`
	AnthropicBaseURL    string `json:"ANTHROPIC_BASE_URL"`
	AnthropicAuthToken  string `json:"ANTHROPIC_AUTH_TOKEN"`
	AnthropicModel      string `json:"ANTHROPIC_MODEL"`
	AnthropicSmallModel string `json:"ANTHROPIC_SMALL_FAST_MODEL"`
}

// Config 表示配置文件结构
type Config struct {
	Platforms []Platform `json:"platforms"`
}

// Validate 验证平台配置是否有效
func (p *Platform) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("平台名称不能为空")
	}
	if p.AnthropicBaseURL == "" {
		return fmt.Errorf("平台 %s 缺少 ANTHROPIC_BASE_URL", p.Name)
	}
	if p.AnthropicAuthToken == "" {
		return fmt.Errorf("平台 %s 缺少 ANTHROPIC_AUTH_TOKEN", p.Name)
	}
	if p.AnthropicModel == "" {
		return fmt.Errorf("平台 %s 缺少 ANTHROPIC_MODEL", p.Name)
	}
	return nil
}

// Validate 验证配置文件是否有效
func (c *Config) Validate() error {
	if len(c.Platforms) == 0 {
		return fmt.Errorf("配置中没有定义任何平台")
	}
	for i, platform := range c.Platforms {
		if err := platform.Validate(); err != nil {
			return fmt.Errorf("平台 %d: %w", i+1, err)
		}
	}
	return nil
}

// getConfigPath 返回默认配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "platforms.json"
	}

	configPath := filepath.Join(homeDir, ".ccgate", "config.json")

	// 如果目录不存在则创建
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0o755)
	}

	return configPath
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	// 确定配置文件路径
	if configPath == "" {
		configPath = getConfigPath()
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Platforms: []Platform{}}, nil
		}
		return nil, fmt.Errorf("无法读取配置文件 %s: %w", configPath, err)
	}

	// 解析 JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		// 尝试直接解析为平台数组（向后兼容）
		var platforms []Platform
		if err2 := json.Unmarshal(data, &platforms); err2 == nil {
			config.Platforms = platforms
		} else {
			return nil, fmt.Errorf("配置文件 JSON 格式无效: %w", err)
		}
	}

	return &config, nil
}

// saveConfig 保存配置到文件
func saveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = getConfigPath()
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("写入配置文件 %s 失败: %w", configPath, err)
	}

	color.Green("✓ 配置已保存到: %s", configPath)
	return nil
}

// findPlatformByName 通过名称查找平台
func findPlatformByName(platforms []Platform, name string) (*Platform, error) {
	for i := range platforms {
		if platforms[i].Name == name {
			return &platforms[i], nil
		}
	}
	return nil, fmt.Errorf("平台 '%s' 不存在", name)
}

// maskToken 掩码敏感令牌信息
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
