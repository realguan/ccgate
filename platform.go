package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

// listPlatforms 列出所有平台
func listPlatforms(platforms []Platform) {
	if len(platforms) == 0 {
		color.Yellow("没有配置任何平台")
		fmt.Println("使用 'ccgate add' 命令添加新平台")
		return
	}

	color.Cyan("\n可用平台 (%d):", len(platforms))
	color.Cyan("========================")
	for i, platform := range platforms {
		color.Green("\n%d. %s", i+1, platform.Name)
		if platform.Vendor != "" {
			fmt.Printf("   厂商: %s\n", platform.Vendor)
		}
		fmt.Printf("   API: %s\n", platform.AnthropicBaseURL)
		fmt.Printf("   模型: %s\n", platform.AnthropicModel)
		if platform.AnthropicSmallModel != "" {
			fmt.Printf("   快速模型: %s\n", platform.AnthropicSmallModel)
		}
	}
	fmt.Println()
}

// addPlatform 交互式添加或更新平台
func addPlatform() (Platform, error) {
	var platform Platform

	color.Cyan("\n添加新的平台配置")
	color.Cyan("==================\n")

	// 平台名称
	namePrompt := promptui.Prompt{
		Label: "平台名称",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("平台名称不能为空")
			}
			return nil
		},
	}
	name, err := namePrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取平台名称失败: %w", err)
	}
	platform.Name = name

	// 厂商（可选）
	vendorPrompt := promptui.Prompt{
		Label: "厂商 (可选)",
	}
	vendor, err := vendorPrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取厂商信息失败: %w", err)
	}
	platform.Vendor = vendor

	// API URL
	urlPrompt := promptui.Prompt{
		Label: "ANTHROPIC_BASE_URL",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("API URL 不能为空")
			}
			return nil
		},
		Default: "https://api.anthropic.com",
	}
	url, err := urlPrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取 API URL 失败: %w", err)
	}
	platform.AnthropicBaseURL = url

	// 认证令牌
	tokenPrompt := promptui.Prompt{
		Label: "ANTHROPIC_AUTH_TOKEN",
		Mask:  '*',
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("认证令牌不能为空")
			}
			return nil
		},
	}
	token, err := tokenPrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取认证令牌失败: %w", err)
	}
	platform.AnthropicAuthToken = token

	// 模型
	modelPrompt := promptui.Prompt{
		Label: "ANTHROPIC_MODEL",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("模型不能为空")
			}
			return nil
		},
		Default: "claude-sonnet-4-20250514",
	}
	model, err := modelPrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取模型失败: %w", err)
	}
	platform.AnthropicModel = model

	// 快速模型（可选）
	fastModelPrompt := promptui.Prompt{
		Label:   "ANTHROPIC_SMALL_FAST_MODEL (可选)",
		Default: "claude-3-5-haiku-20241022",
	}
	fastModel, err := fastModelPrompt.Run()
	if err != nil {
		return platform, fmt.Errorf("获取快速模型失败: %w", err)
	}
	platform.AnthropicSmallModel = fastModel

	// 验证配置
	if err := platform.Validate(); err != nil {
		return platform, fmt.Errorf("平台配置验证失败: %w", err)
	}

	color.Green("\n✓ 平台配置验证通过")
	return platform, nil
}

// deletePlatform 删除指定名称的平台
func deletePlatform(platforms []Platform, name string) ([]Platform, error) {
	for i, platform := range platforms {
		if platform.Name == name {
			return append(platforms[:i], platforms[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("平台 '%s' 不存在", name)
}

// updateOrAddPlatform 更新或添加平台
func updateOrAddPlatform(platforms []Platform, newPlatform Platform) []Platform {
	for i, p := range platforms {
		if p.Name == newPlatform.Name {
			platforms[i] = newPlatform
			return platforms
		}
	}
	return append(platforms, newPlatform)
}
