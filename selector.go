package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// selectPlatform 选择平台（自动或交互式）
// claudeArgs 用于判断是否需要显示提示信息
func selectPlatform(config *Config, platformName string, claudeArgs []string) (*Platform, error) {
	if len(config.Platforms) == 0 {
		return nil, fmt.Errorf("没有配置任何平台\n请先运行 'ccgate add' 添加平台")
	}

	// 情况1: 通过 -p/--platform 显式指定
	if platformName != "" {
		platform, err := findPlatformByName(config.Platforms, platformName)
		if err != nil {
			// 提供模糊匹配建议
			suggestions := suggestPlatformNames(config.Platforms, platformName)
			if len(suggestions) > 0 {
				return nil, fmt.Errorf(
					"平台 '%s' 不存在\n\n你是否想使用以下平台？\n  - %s\n\n运行 'ccgate list' 查看所有可用平台",
					platformName,
					strings.Join(suggestions, "\n  - "),
				)
			}
			return nil, fmt.Errorf("%w\n运行 'ccgate list' 查看所有可用平台", err)
		}
		return platform, nil
	}

	// 情况2: 只有一个平台，自动使用
	if len(config.Platforms) == 1 {
		platform := &config.Platforms[0]
		if len(claudeArgs) > 0 {
			color.Cyan("→ 检测到唯一平台: %s，自动使用", platform.Name)
		}
		return platform, nil
	}

	// 情况3: 多个平台，需要交互式选择
	// 检查是否支持交互（TTY）
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, formatNonInteractiveError(config.Platforms, claudeArgs)
	}

	// 显示提示信息
	if len(claudeArgs) > 0 {
		color.Yellow("\n→ 检测到 claude 命令参数: %s", strings.Join(claudeArgs, " "))
		fmt.Println()
	}

	// 交互式选择
	platform, err := interactiveSelectPlatform(config.Platforms)
	if err != nil {
		return nil, err
	}

	return platform, nil
}

// interactiveSelectPlatform 交互式选择平台
func interactiveSelectPlatform(platforms []Platform) (*Platform, error) {
	// 构建选择模板
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Name | cyan }} {{ if .Vendor }}({{ .Vendor | faint }}){{ end }}",
		Inactive: "  {{ .Name }} {{ if .Vendor }}({{ .Vendor | faint }}){{ end }}",
		Selected: color.GreenString("→ 已选择: {{ .Name }}"),
		Details: `
--------- 平台详情 ---------
{{ "名称:" | faint }}	{{ .Name }}
{{ "厂商:" | faint }}	{{ .Vendor }}
{{ "Base URL:" | faint }}	{{ .ANTHROPIC_BASE_URL }}
{{ "模型:" | faint }}	{{ .ANTHROPIC_MODEL }}`,
	}

	// 创建选择器
	prompt := promptui.Select{
		Label:     "选择平台",
		Items:     platforms,
		Templates: templates,
		Size:      10,
		// 支持搜索
		Searcher: func(input string, index int) bool {
			platform := platforms[index]
			name := strings.ToLower(platform.Name)
			input = strings.ToLower(input)
			return strings.Contains(name, input)
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		// 用户取消选择
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF || err == promptui.ErrAbort {
			color.Yellow("\n平台选择已取消")
			os.Exit(0)
		}
		return nil, fmt.Errorf("平台选择失败: %w", err)
	}

	return &platforms[idx], nil
}

// confirmExecution 确认执行
func confirmExecution(platform *Platform, claudeArgs []string, skipConfirm bool) error {
	if skipConfirm {
		return nil
	}

	fmt.Println()
	color.Green("→ 将使用平台: %s", platform.Name)
	if platform.Vendor != "" {
		color.Cyan("  厂商: %s", platform.Vendor)
	}
	color.Cyan("  Base URL: %s", platform.AnthropicBaseURL)
	color.Cyan("  模型: %s", platform.AnthropicModel)
	color.Yellow("  认证令牌: %s", maskToken(platform.AnthropicAuthToken))

	if len(claudeArgs) > 0 {
		color.Magenta("\n→ 执行命令: claude %s", strings.Join(claudeArgs, " "))
	} else {
		color.Magenta("\n→ 执行命令: claude (交互式)")
	}

	fmt.Println()

	prompt := promptui.Prompt{
		Label:     "确认执行",
		IsConfirm: true,
		Default:   "Y",
	}

	result, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("操作已取消")
	}

	// 只接受 Y/y/回车
	if result != "" && result != "y" && result != "Y" {
		return fmt.Errorf("操作已取消")
	}

	return nil
}

// formatNonInteractiveError 格式化非交互环境错误信息
func formatNonInteractiveError(platforms []Platform, claudeArgs []string) error {
	names := make([]string, len(platforms))
	for i, p := range platforms {
		names[i] = fmt.Sprintf("  - %s", p.Name)
	}

	cmdExample := "ccgate -p <平台名>"
	if len(claudeArgs) > 0 {
		cmdExample = fmt.Sprintf("ccgate -p <平台名> %s", strings.Join(claudeArgs, " "))
	}

	return fmt.Errorf(
		"错误：检测到 %d 个平台，但当前环境不支持交互式选择\n\n"+
			"可用平台:\n%s\n\n"+
			"请使用 -p/--platform 显式指定平台:\n  %s\n\n"+
			"示例:\n"+
			"  ccgate -p production --continue\n"+
			"  ccgate -p staging chat \"hello\"",
		len(platforms),
		strings.Join(names, "\n"),
		cmdExample,
	)
}

// suggestPlatformNames 基于 Levenshtein 距离返回相似的平台名称
func suggestPlatformNames(platforms []Platform, query string) []string {
	type candidate struct {
		name string
		dist int
	}

	var candidates []candidate
	for _, p := range platforms {
		d := levenshteinDistance(p.Name, query)
		candidates = append(candidates, candidate{name: p.Name, dist: d})
	}

	// 按距离排序
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].dist < candidates[i].dist {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// 返回最多 5 个建议
	var results []string
	for i := 0; i < len(candidates) && i < 5; i++ {
		results = append(results, candidates[i].name)
	}
	return results
}

// levenshteinDistance 计算两个字符串的 Levenshtein 距离
func levenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
