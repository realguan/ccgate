package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Platform 定义了 Claude Code 需要的环境变量配置
// 通过配置 Base URL，可以适配 MiniMax, GLM (智谱), Moonshot (Kimi) 等支持兼容接口的服务
type Platform struct {
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"` // 描述，例如 "MiniMax abab6.5"
	Vendor              string `json:"vendor,omitempty"`      // 厂商，例如 "MiniMax", "ZhipuAI"
	Pinned              bool   `json:"pinned,omitempty"`      // 是否置顶
	
	// 核心配置
	AnthropicBaseURL    string `json:"ANTHROPIC_BASE_URL"`
	AnthropicAuthToken  string `json:"ANTHROPIC_AUTH_TOKEN"`
	AnthropicModel      string `json:"ANTHROPIC_MODEL"`
	AnthropicSmallModel string `json:"ANTHROPIC_SMALL_FAST_MODEL,omitempty"`
	
	// 扩展配置 (备用，未来可能支持更多自定义 env)
	ExtraEnv            map[string]string `json:"extra_env,omitempty"`
}

// Config 是顶层配置结构
type Config struct {
	Platforms []Platform `json:"platforms"`
}

// FilterValue 用于 bubbles/list 的模糊搜索
func (p Platform) FilterValue() string {
	return p.Name + " " + p.Vendor + " " + p.Description
}

// Title 用于 bubbles/list 的主标题
func (p Platform) Title() string {
	if p.Pinned {
		return "⭐ " + p.Name
	}
	return p.Name
}

// DescriptionText 用于 bubbles/list 的副标题
func (p Platform) DescriptionText() string {
	if p.Description != "" {
		return p.Description
	}
	if p.Vendor != "" {
		return fmt.Sprintf("%s | %s", p.Vendor, p.AnthropicModel)
	}
	return p.AnthropicModel
}

// MaskedToken 返回脱敏后的 Token
func (p Platform) MaskedToken() string {
	if len(p.AnthropicAuthToken) <= 8 {
		return "********"
	}
	return p.AnthropicAuthToken[:4] + "****" + p.AnthropicAuthToken[len(p.AnthropicAuthToken)-4:]
}

// GenerateExampleConfig 生成包含 MiniMax, GLM 等示例的默认配置
func GenerateExampleConfig() *Config {
	return &Config{
		Platforms: []Platform{
			{
				Name:               "anthropic-official",
				Vendor:             "Anthropic",
				Description:        "Claude 官方接口",
				AnthropicBaseURL:   "https://api.anthropic.com",
				AnthropicAuthToken: "sk-ant-...",
				AnthropicModel:     "claude-3-5-sonnet-20240620",
				Pinned:             true,
			},
			{
				Name:               "minimax",
				Vendor:             "MiniMax",
				Description:        "MiniMax 通用大模型",
				AnthropicBaseURL:   "https://api.minimax.chat/v1", // 示例地址，需根据实际兼容网关调整
				AnthropicAuthToken: "your-minimax-api-key",
				AnthropicModel:     "abab6.5-chat",
			},
			{
				Name:               "zhipu-glm",
				Vendor:             "ZhipuAI",
				Description:        "智谱 GLM-4",
				AnthropicBaseURL:   "https://open.bigmodel.cn/api/paas/v4", // 示例地址
				AnthropicAuthToken: "your-zhipu-api-key",
				AnthropicModel:     "glm-4",
			},
		},
	}
}

// LoadConfig 加载配置，优先查找 ~/.cc-proxy/config.json，其次 ~/.ccgate/config.json
func LoadConfig() (*Config, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, "", err
	}

	// 路径优先级
	paths := []string{
		filepath.Join(home, ".cc-proxy", "config.json"),
		filepath.Join(home, ".ccgate", "config.json"),
	}

	var configPath string
	var data []byte

	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			configPath = p
			data = d
			break
		}
	}

	// 如果都没找到，使用默认路径（虽然文件不存在）
	if configPath == "" {
		configPath = paths[0]
		return &Config{Platforms: []Platform{}}, configPath, nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		// 尝试兼容旧格式（根数组）
		var platforms []Platform
		if err2 := json.Unmarshal(data, &platforms); err2 == nil {
			config.Platforms = platforms
		} else {
			return nil, configPath, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	return &config, configPath, nil
}

// SaveConfig 保存配置
func SaveConfig(config *Config, path string) error {
	if path == "" {
		// 默认路径
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".cc-proxy", "config.json")
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}