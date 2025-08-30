package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"golang.org/x/term"
)

// Platform represents a platform configuration
type Platform struct {
	Name                string `json:"name"`
	Vendor              string `json:"vendor"`
	AnthropicBaseURL    string `json:"ANTHROPIC_BASE_URL"`
	AnthropicAuthToken  string `json:"ANTHROPIC_AUTH_TOKEN"`
	AnthropicModel      string `json:"ANTHROPIC_MODEL"`
	AnthropicSmallModel string `json:"ANTHROPIC_SMALL_FAST_MODEL"`
}

// Config represents the configuration structure
type Config struct {
	Platforms []Platform `json:"platforms"`
}

// validatePlatforms checks all platforms and returns an error if any are invalid
func validatePlatforms(platforms []Platform) error {
	if len(platforms) == 0 {
		return fmt.Errorf("no platforms defined in configuration")
	}
	for i, platform := range platforms {
		if err := platform.Validate(); err != nil {
			return fmt.Errorf("platform %d: %v", i+1, err)
		}
	}
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	return validatePlatforms(c.Platforms)
}

// Validate checks if the platform configuration is valid
func (p *Platform) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("platform name is required")
	}
	// Vendor is optional, so no validation required
	if p.AnthropicBaseURL == "" {
		return fmt.Errorf("ANTHROPIC_BASE_URL is required for platform %s", p.Name)
	}
	if p.AnthropicAuthToken == "" {
		return fmt.Errorf("ANTHROPIC_AUTH_TOKEN is required for platform %s", p.Name)
	}
	if p.AnthropicModel == "" {
		return fmt.Errorf("ANTHROPIC_MODEL is required for platform %s", p.Name)
	}
	return nil
}

func main() {
	// 只保留 Cobra CLI 入口
	Execute()
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home directory cannot be determined
		return "platforms.json"
	}

	// Default config path
	configPath := filepath.Join(homeDir, ".cctool", "config.json")

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0o755)
	}

	return configPath
}

func loadPlatforms(configFilePath string) ([]Platform, error) {
	var data []byte
	var err error

	// If config file path is specified via -f flag, use it
	if configFilePath != "" {
		data, err = os.ReadFile(configFilePath)
		if err != nil {
			return nil, fmt.Errorf("无法读取配置文件 %s: %v", configFilePath, err)
		}
		color.Cyan("已加载配置文件: %s", configFilePath)
	} else {
		// Try to read from default config path first
		configPath := getConfigPath()
		// Add debug info
		color.Magenta("调试: 尝试加载配置文件: %s", configPath)
		data, err = os.ReadFile(configPath)
		if err != nil {
			// If no config file found, guide user to create one immediately
			color.Yellow("未找到配置文件。")
			color.Yellow("请使用以下方式之一来配置平台:")
			color.Yellow("  1. 运行 'cctool -add' 命令添加新平台")
			color.Yellow("  2. 创建配置文件 ~/.cctool/config.json")
			return []Platform{}, nil
		} else {
			color.Cyan("已加载配置文件: %s", configPath)
		}
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		// Try to parse as array of platforms directly
		var platforms []Platform
		err2 := json.Unmarshal(data, &platforms)
		if err2 != nil {
			return nil, fmt.Errorf("配置文件中的JSON格式无效: %v", err)
		}
		config.Platforms = platforms
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		// If config is empty, guide user to add platforms
		if len(config.Platforms) == 0 {
			color.Yellow("配置文件中没有定义任何平台。")
			color.Yellow("请使用 'cctool -add' 命令添加新平台")
			return config.Platforms, nil
		}
		return nil, fmt.Errorf("配置无效: %v", err)
	}

	color.Green("成功加载 %d 个平台配置", len(config.Platforms))
	return config.Platforms, nil
}

func selectPlatformInteractive(platforms []Platform) (Platform, error) {
	// Check if there are any platforms available
	if len(platforms) == 0 {
		return Platform{}, fmt.Errorf("没有可用的平台配置，请先添加平台")
	}

	// Create a slice of platform names for the selector
	names := make([]string, len(platforms))
	for i, platform := range platforms {
		names[i] = platform.Name
	}



	// Create the prompt
	prompt := promptui.Select{
		Label: "选择平台 (Ctrl+C退出)",
		Items: names,
	}

	// Run the prompt
	_, result, err := prompt.Run()
	if err != nil {
		// Check if the user pressed ESC or Ctrl+C to exit
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF || err == promptui.ErrAbort || err.Error() == "Interrupt" {
			color.Yellow("用户取消选择，退出程序。")
			os.Exit(0)
		}
		return Platform{}, fmt.Errorf("提示失败: %v", err)
	}

	// Find the selected platform
	for _, platform := range platforms {
		if platform.Name == result {
			return platform, nil
		}
	}

	// This should never happen
	return Platform{}, fmt.Errorf("未找到选定的平台: %s", result)
}

func setEnvironment(platform Platform) {
	os.Setenv("ANTHROPIC_BASE_URL", platform.AnthropicBaseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", platform.AnthropicAuthToken)
	os.Setenv("ANTHROPIC_MODEL", platform.AnthropicModel)
	os.Setenv("ANTHROPIC_SMALL_FAST_MODEL", platform.AnthropicSmallModel)
}

func launchClaudeCode() {
	// This will launch Claude Code with the exported environment variables
	cmd := exec.Command("claude")
	cmd.Env = os.Environ()
	
	// 检测是否在交互式终端中运行
	isInteractive := term.IsTerminal(int(os.Stdin.Fd()))
	
	if isInteractive {
		// 在交互式终端中，连接 stdin/stdout/stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 在非交互式环境中，断开 stdin/stdout/stderr 连接
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	color.Cyan("执行命令: claude")
	err := cmd.Start()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			color.Red("错误: 未找到 Claude Code 命令。请确保它已安装并在您的 PATH 中。")
			color.Yellow("您可以通过运行以下命令手动启动: claude")
			showInstallationInstructions()
		} else {
			color.Red("启动 Claude Code 时出错: %v", err)
			color.Yellow("您可以通过运行以下命令手动启动: claude")
		}
		return
	}

	if isInteractive {
		// 等待进程结束（交互模式）
		err = cmd.Wait()
		if err != nil {
			color.Red("Claude Code 运行时出错: %v", err)
		} else {
			color.Green("Claude Code 已退出。")
		}
	} else {
		// 立即退出，让 claude 在后台继续运行（非交互模式）
		color.Green("Claude Code 已在后台启动!")
	}
}

// launchClaudeWithPlatform handles the platform selection and launching of Claude Code
func launchClaudeWithPlatform(platforms []Platform, platformFlag string, _ string) error {
	// Use specified platform if provided
	var selectedPlatform Platform
	var err error
	if platformFlag != "" {
		selectedPlatform, err = findPlatformByName(platforms, platformFlag)
		if err != nil {
			return err
		}
		color.Green("\n使用平台: %s\n", selectedPlatform.Name)
	} else {
		// Display platforms and get user choice with interactive menu
		selectedPlatform, err = selectPlatformInteractive(platforms)
		if err != nil {
			return err
		}
		color.Green("\n已选择平台: %s\n", selectedPlatform.Name)
	}

	// Show platform details before confirmation
	showPlatformDetails(selectedPlatform)

	// Check if we should skip confirmation
	if !skipConfirm {
		// Confirmation loop: prompt user to confirm launch, allow re-selection
		for {
			// Confirm before launching
			if confirmLaunch() {
				break // User confirmed, proceed with launch
			}

			color.Yellow("启动已取消。")
			// Instead of exiting, ask if user wants to select another platform
			prompt := promptui.Prompt{
				Label:     "是否选择其他平台",
				IsConfirm: true,
				Default:   "y",
			}
			result, err := prompt.Run()
			if err != nil || (result != "" && result != "y" && result != "Y") {
				return nil // User chose not to select another platform or there was an error
			}

			// Go back to platform selection
			selectedPlatform, err = selectPlatformInteractive(platforms)
			if err != nil {
				return err
			}
			color.Green("\n已选择平台: %s\n", selectedPlatform.Name)
		}
	}

	// proceed to set environment and launch

	// Set environment variables
	setEnvironment(selectedPlatform)
	color.Green("环境变量设置成功!")

	// Launch Claude Code
	color.Cyan("正在启动 Claude Code...")
	
	// 检测是否在交互式终端中运行
	isInteractive := term.IsTerminal(int(os.Stdin.Fd()))
	
	launchClaudeCode()
	
	if !isInteractive {
		// 在非交互模式下，立即退出，让 claude 在后台继续运行
		color.Green("cctool 任务完成，已退出。Claude Code 在后台运行中...")
		os.Exit(0)
	}
	
	return nil
}

// showInstallationInstructions provides guidance on how to install Claude Code
func showInstallationInstructions() {
	color.Cyan("\n安装说明:")
	color.Cyan("==========================")
	color.Yellow("要安装 Claude Code，请访问: https://claude.ai/download")
	fmt.Println("或按照以下步骤操作:")
	color.Green("1. 从官方网站下载 Claude Code 应用程序")
	color.Green("2. 按照提供的说明安装应用程序")
	color.Green("3. 确保 'claude' 命令在您的 PATH 中")
	color.Green("4. 再次运行此工具")
}

// deletePlatform removes a platform by name
func deletePlatform(platforms []Platform, name string) ([]Platform, error) {
	for i, platform := range platforms {
		if platform.Name == name {
			// Remove the platform at index i
			return append(platforms[:i], platforms[i+1:]...), nil
		}
	}
	return nil, fmt.Errorf("platform '%s' not found", name)
}

// addNewPlatform prompts user to add a new platform
func addNewPlatform() (Platform, error) {
	var newPlatform Platform

	color.Cyan("添加新的平台配置:")
	color.Cyan("==================")

	// Prompt for platform name
	namePrompt := promptui.Prompt{
		Label: "name",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("平台名称不能为空")
			}
			return nil
		},
	}
	name, err := namePrompt.Run()
	if err != nil {
		return newPlatform, fmt.Errorf("获取平台名称失败: %v", err)
	}
	newPlatform.Name = name

	// Prompt for vendor
	vendorPrompt := promptui.Prompt{
		Label: "vendor (optional)",
	}
	vendor, err := vendorPrompt.Run()
	if err != nil {
		return newPlatform, fmt.Errorf("获取厂商信息失败: %v", err)
	}
	newPlatform.Vendor = vendor

	// Prompt for API URL
	urlPrompt := promptui.Prompt{
		Label: "ANTHROPIC_BASE_URL",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("API URL 不能为空")
			}
			return nil
		},
	}
	url, err := urlPrompt.Run()
	if err != nil {
		return newPlatform, fmt.Errorf("获取 API URL 失败: %v", err)
	}
	newPlatform.AnthropicBaseURL = url

	// Prompt for Auth Token
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
		return newPlatform, fmt.Errorf("获取认证令牌失败: %v", err)
	}
	newPlatform.AnthropicAuthToken = token

	// Prompt for Model
	modelPrompt := promptui.Prompt{
		Label: "ANTHROPIC_MODEL",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("模型不能为空")
			}
			return nil
		},
	}
	model, err := modelPrompt.Run()
	if err != nil {
		return newPlatform, fmt.Errorf("获取模型失败: %v", err)
	}
	newPlatform.AnthropicModel = model

	// Prompt for Fast Model
	fastModelPrompt := promptui.Prompt{
		Label: "ANTHROPIC_SMALL_FAST_MODEL",
		Validate: func(input string) error {
			if input == "" {
				return fmt.Errorf("快速模型不能为空")
			}
			return nil
		},
	}
	fastModel, err := fastModelPrompt.Run()
	if err != nil {
		return newPlatform, fmt.Errorf("获取快速模型失败: %v", err)
	}
	newPlatform.AnthropicSmallModel = fastModel

	// Validate the new platform
	if err := newPlatform.Validate(); err != nil {
		return newPlatform, fmt.Errorf("无效的平台配置: %v", err)
	}

	color.Green("平台配置验证通过!")
	return newPlatform, nil
}

// savePlatforms saves the platforms to the config file
func savePlatforms(platforms []Platform, configFilePath string) error {
	config := Config{
		Platforms: platforms,
	}

	// Convert to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化平台配置失败: %v", err)
	}

	// Save to config file
	var path string
	if configFilePath != "" {
		path = configFilePath
	} else {
		path = getConfigPath()
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %v", path, err)
	}

	color.Green("配置已保存到: %s", path)
	return nil
}

// listPlatforms displays all available platforms
func listPlatforms(platforms []Platform) {
	if len(platforms) == 0 {
		color.Yellow("没有配置任何平台。")
		fmt.Println("使用 'cctool -add' 命令添加新平台。")
		return
	}

	color.Cyan("可用平台 (%d):", len(platforms))
	color.Cyan("========================")
	for i, platform := range platforms {
		color.Green("%d. %s", i+1, platform.Name)
		// Show vendor if available
		if platform.Vendor != "" {
			fmt.Printf("   厂商: %s\n", platform.Vendor)
		}
		// Show abbreviated URL
		if len(platform.AnthropicBaseURL) > 50 {
			fmt.Printf("   API: %s...\n", platform.AnthropicBaseURL[:50])
		} else {
			fmt.Printf("   API: %s\n", platform.AnthropicBaseURL)
		}
		fmt.Printf("   模型: %s\n", platform.AnthropicModel)
		fmt.Println()
	}
}

// findPlatformByName finds a platform by its name
func findPlatformByName(platforms []Platform, name string) (Platform, error) {
	for _, platform := range platforms {
		if platform.Name == name {
			return platform, nil
		}
	}
	return Platform{}, fmt.Errorf("platform '%s' not found", name)
}

// showPlatformDetails displays detailed information about the selected platform
func showPlatformDetails(platform Platform) {
	color.Cyan("\n平台详情:")
	color.Cyan("=================")
	color.Green("名称: %s", platform.Name)

	// Show vendor if available
	if platform.Vendor != "" {
		fmt.Printf("厂商: %s\n", platform.Vendor)
	}

	// Show full URL
	fmt.Printf("API URL: %s\n", platform.AnthropicBaseURL)

	color.Green("模型: %s", platform.AnthropicModel)
	color.Green("快速模型: %s", platform.AnthropicSmallModel)

	// Show token masked
	if len(platform.AnthropicAuthToken) > 8 {
		color.Yellow("认证令牌: %s****%s", platform.AnthropicAuthToken[:4], platform.AnthropicAuthToken[len(platform.AnthropicAuthToken)-4:])
	} else {
		color.Yellow("认证令牌: ****")
	}
}

// confirmLaunch asks user to confirm before launching
func confirmLaunch() bool {
	prompt := promptui.Prompt{
		Label:     "是否使用此配置启动 Claude Code",
		IsConfirm: true,
		Default:   "y",
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	return result == "" || result == "y" || result == "Y"
}

// maskToken masks sensitive parts of a token for safe display
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

// suggestPlatformNames returns up to 5 platform names similar to query using Levenshtein distance
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

	// sort by dist ascending
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].dist < candidates[i].dist {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	var results []string
	for i := 0; i < len(candidates) && i < 5; i++ {
		results = append(results, candidates[i].name)
	}
	return results
}

// levenshteinDistance computes the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}

	for i := 0; i <= la; i++ {
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
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
