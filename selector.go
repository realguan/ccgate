package main

import (
	"fmt"
	"os"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

// selectPlatform 选择平台（自动或交互式）
// claudeArgs 用于判断是否需要显示提示信息
// skipConfirm 是否跳过确认（用于 --yes 参数）
func selectPlatform(config *Config, platformName string, claudeArgs []string, skipConfirm bool) (*Platform, error) {
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

	// 循环选择，支持 ESC 返回重新选择
	for {
		// 交互式选择（提示信息在函数内部显示）
		platform, err := interactiveSelectPlatform(config.Platforms, claudeArgs)
		if err != nil {
			return nil, err
		}

		// 跳过确认（--yes 参数）
		if skipConfirm {
			return platform, nil
		}

		// 确认执行，如果取消则返回重新选择
		err = confirmExecution(platform, claudeArgs, skipConfirm)
		if err != nil {
			// 用户取消确认，清屏后重新选择
			fmt.Print("\033[H\033[2J")
			pterm.Warning.Println("已取消，重新选择平台")
			fmt.Println()
			continue
		}

		// 确认通过，返回选中的平台
		return platform, nil
	}
}

// interactiveSelectPlatform 交互式选择平台
func interactiveSelectPlatform(platforms []Platform, claudeArgs []string) (*Platform, error) {
	// 显示提示信息
	if len(claudeArgs) > 0 {
		pterm.Warning.Printf("检测到 claude 命令参数: %s\n\n", strings.Join(claudeArgs, " "))
	}

	// 构建选项列表（包含详细信息）
	options := make([]string, len(platforms))
	optionDetails := make([]string, len(platforms))

	for i, p := range platforms {
		// 主显示：平台名称和厂商
		if p.Vendor != "" {
			options[i] = fmt.Sprintf("%s (%s)", p.Name, p.Vendor)
		} else {
			options[i] = p.Name
		}

		// 详细信息（用于搜索和显示）
		optionDetails[i] = fmt.Sprintf("%s %s %s %s",
			p.Name,
			p.Vendor,
			p.AnthropicBaseURL,
			p.AnthropicModel,
		)
	}

	// 创建交互式选择器
	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultText("选择平台 (↑↓ 导航, / 搜索, Enter 确认)").
		WithFilter(true).  // 启用模糊搜索
		WithMaxHeight(15). // 最大显示高度
		Show()

	if err != nil {
		// 用户取消选择 (Ctrl+C)
		pterm.Warning.Println("\n平台选择已取消")
		os.Exit(0)
	}

	// 找到选中的平台索引
	selectedIndex := -1
	for i, opt := range options {
		if opt == selectedOption {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		return nil, fmt.Errorf("未找到选中的平台")
	}

	platform := &platforms[selectedIndex]

	// 显示选中平台的详细信息
	fmt.Println()
	pterm.DefaultSection.Println("平台详情")

	// 构建详情表格
	tableData := pterm.TableData{
		{"名称", platform.Name},
		{"厂商", platform.Vendor},
		{"Base URL", platform.AnthropicBaseURL},
		{"模型", platform.AnthropicModel},
	}

	if platform.AnthropicSmallModel != "" {
		tableData = append(tableData, []string{"快速模型", platform.AnthropicSmallModel})
	}

	// 渲染表格
	pterm.DefaultTable.WithHasHeader(false).
		WithBoxed(true).
		WithData(tableData).
		Render()

	fmt.Println()

	return platform, nil
}

// confirmExecution 确认执行，支持 ESC 键直接返回
func confirmExecution(platform *Platform, claudeArgs []string, skipConfirm bool) error {
	if skipConfirm {
		return nil
	}

	// 显示执行命令
	if len(claudeArgs) > 0 {
		pterm.Info.Printf("执行命令: %s\n\n", pterm.LightMagenta("claude "+strings.Join(claudeArgs, " ")))
	} else {
		pterm.Info.Printf("执行命令: %s\n\n", pterm.LightMagenta("claude (交互式)"))
	}

	// 显示提示
	fmt.Printf("确认执行? [Y/n] (ESC 返回): ")

	// 使用 keyboard 库监听按键
	confirmed := false
	cancelled := false

	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Enter:
			// Enter 键 - 确认
			confirmed = true
			fmt.Println("Y")
			return true, nil

		case keys.RuneKey:
			// 字符键
			if len(key.Runes) > 0 {
				char := string(key.Runes)
				switch char {
				case "y", "Y":
					// Y 键 - 确认
					confirmed = true
					fmt.Println(char)
					return true, nil
				case "n", "N":
					// N 键 - 取消
					cancelled = true
					fmt.Println(char)
					return true, nil
				}
			}

		case keys.Escape:
			// ESC 键 - 取消
			cancelled = true
			fmt.Println("ESC")
			return true, nil

		case keys.CtrlC:
			// Ctrl+C - 退出程序
			fmt.Println()
			pterm.Warning.Println("操作已取消")
			os.Exit(0)
		}

		// 忽略其他按键，继续监听
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("键盘监听失败: %w", err)
	}

	// 根据用户选择返回结果
	if cancelled {
		return fmt.Errorf("操作已取消")
	}

	if !confirmed {
		// 理论上不会到这里，因为 Listen 会一直等待直到 stop=true
		return fmt.Errorf("未确认执行")
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
