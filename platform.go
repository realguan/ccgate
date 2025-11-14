package main

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// listPlatforms 列出所有平台
func listPlatforms(platforms []Platform) {
	theme := DefaultTheme()

	if len(platforms) == 0 {
		DisplayWarning("没有配置任何平台", theme)
		fmt.Println("使用 'ccgate add' 命令添加新平台")
		return
	}

	// 显示标题，使用主题主色
	title := fmt.Sprintf("可用平台 (%d)", len(platforms))
	pterm.Info.Printf("%s\n", theme.Colors.Primary.Sprint(title))
	pterm.Info.Printf("%s\n", theme.Colors.Secondary.Sprint(strings.Repeat("=", len(title))))

	for i, platform := range platforms {
		Spacer(theme.Spacing.SM, theme)

		// 平台编号和名称
		pterm.Printf("%s %s\n",
			theme.Colors.Success.Sprint(fmt.Sprintf("%d.", i+1)),
			theme.Colors.Primary.Sprint(platform.Name))

		// 平台详情
		if platform.Vendor != "" {
			pterm.Printf("   %s %s\n",
				theme.Colors.Secondary.Sprint("厂商:"),
				platform.Vendor)
		}
		pterm.Printf("   %s %s\n",
			theme.Colors.Secondary.Sprint("API:"),
			platform.AnthropicBaseURL)
		pterm.Printf("   %s %s\n",
			theme.Colors.Secondary.Sprint("模型:"),
			platform.AnthropicModel)
		if platform.AnthropicSmallModel != "" {
			pterm.Printf("   %s %s\n",
				theme.Colors.Secondary.Sprint("快速模型:"),
				theme.Colors.Info.Sprint(platform.AnthropicSmallModel))
		}
	}

	Spacer(theme.Spacing.MD, theme)
}

// addPlatform 交互式添加或更新平台
func addPlatform() (Platform, error) {
	var platform Platform
	theme := DefaultTheme()

	// 显示标题
	pterm.Info.Printf("%s\n", theme.Colors.Primary.Sprint("🚀 添加新的平台配置"))
	pterm.Info.Printf("%s\n\n", theme.Colors.Secondary.Sprint(strings.Repeat("=", 20)))

	// 平台名称
	pterm.Printf("%s\n", theme.Colors.Primary.Sprint("📝 平台名称"))
	for {
		name, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("请输入平台名称（如：production, staging, development）").
			Show()
		if err != nil {
			return platform, fmt.Errorf("获取平台名称失败: %w", err)
		}

		// 验证
		if strings.TrimSpace(name) == "" {
			err := NewValidationError("平台名称不能为空", "请输入一个有效的平台名称")
			err.DisplayError(theme)
			continue
		}

		platform.Name = strings.TrimSpace(name)
		break
	}

	// 厂商（可选）
	pterm.Printf("\n%s\n", theme.Colors.Secondary.Sprint("🏢 厂商（可选）"))
	vendor, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("请输入厂商名称（如：Anthropic, OpenAI, 第三方代理商）").
		Show()
	if err != nil {
		return platform, fmt.Errorf("获取厂商信息失败: %w", err)
	}
	platform.Vendor = strings.TrimSpace(vendor)

	// API URL
	pterm.Printf("\n%s\n", theme.Colors.Primary.Sprint("🔗 ANTHROPIC_BASE_URL"))
	for {
		url, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("请输入 API Base URL（如：https://api.anthropic.com）").
			Show()
		if err != nil {
			return platform, fmt.Errorf("获取 API URL 失败: %w", err)
		}

		// 验证
		if strings.TrimSpace(url) == "" {
			err := NewValidationError("API URL 不能为空", "请输入有效的 API URL")
			err.DisplayError(theme)
			continue
		}

		platform.AnthropicBaseURL = strings.TrimSpace(url)
		break
	}

	// 认证令牌
	pterm.Printf("\n%s\n", theme.Colors.Primary.Sprint("🔑 ANTHROPIC_AUTH_TOKEN"))
	for {
		token, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("请输入认证令牌（API Key）").
			Show()
		if err != nil {
			return platform, fmt.Errorf("获取认证令牌失败: %w", err)
		}

		// 验证
		if strings.TrimSpace(token) == "" {
			err := NewValidationError("认证令牌不能为空", "请输入有效的认证令牌")
			err.DisplayError(theme)
			continue
		}

		platform.AnthropicAuthToken = strings.TrimSpace(token)
		break
	}

	// 模型
	pterm.Printf("\n%s\n", theme.Colors.Primary.Sprint("🤖 ANTHROPIC_MODEL"))
	for {
		model, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("请输入模型名称（如：claude-sonnet-4-20250514）").
			Show()
		if err != nil {
			return platform, fmt.Errorf("获取模型失败: %w", err)
		}

		// 验证
		if strings.TrimSpace(model) == "" {
			err := NewValidationError("模型不能为空", "请输入有效的模型名称")
			err.DisplayError(theme)
			continue
		}

		platform.AnthropicModel = strings.TrimSpace(model)
		break
	}

	// 快速模型（可选）
	pterm.Printf("\n%s\n", theme.Colors.Secondary.Sprint("⚡ ANTHROPIC_SMALL_FAST_MODEL（可选）"))
	fastModel, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("请输入快速模型名称（如：claude-3-5-haiku-20241022，回车跳过）").
		Show()
	if err != nil {
		return platform, fmt.Errorf("获取快速模型失败: %w", err)
	}
	platform.AnthropicSmallModel = strings.TrimSpace(fastModel)

	Spacer(theme.Spacing.MD, theme)

	// 验证配置
	if err := platform.Validate(); err != nil {
		return platform, fmt.Errorf("平台配置验证失败: %w", err)
	}

	DisplaySuccess("✓ 平台配置验证通过", theme)
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
