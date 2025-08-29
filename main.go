package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
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

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Platforms) == 0 {
		return fmt.Errorf("no platforms defined in configuration")
	}

	for i, platform := range c.Platforms {
		if err := platform.Validate(); err != nil {
			return fmt.Errorf("platform %d: %v", i+1, err)
		}
	}

	return nil
}

func main() {
	// Define command line flags
	listFlag := flag.Bool("list", false, "List all available platforms")
	platformFlag := flag.String("platform", "", "Specify platform to use directly")
	configFlag := flag.String("f", "", "Specify config file path")
	helpFlag := flag.Bool("help", false, "Show help message")
	hFlag := flag.Bool("h", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version information")
	vFlag := flag.Bool("v", false, "Show version information")
	addFlag := flag.Bool("add", false, "Add a new platform")
	deleteFlag := flag.String("delete", "", "Delete a platform by name")

	// Parse flags
	flag.Parse()

	// Handle command line flags
	err := handleCommandLineFlags(
		*listFlag, *platformFlag, *configFlag, *helpFlag, *hFlag,
		*versionFlag, *vFlag, *addFlag, *deleteFlag,
	)
	if err != nil {
		color.Red("错误: %v\n", err)
		os.Exit(1)
	}
}

// handleCommandLineFlags processes command line flags and executes corresponding actions
func handleCommandLineFlags(
	listFlag bool, platformFlag string, configFlag string, helpFlag bool, hFlag bool,
	versionFlag bool, vFlag bool, addFlag bool, deleteFlag string,
) error {
	// Show version if requested
	if versionFlag || vFlag {
		color.Cyan("Claude Code 平台选择器 v1.1.0")
		return nil
	}

	// Show help if requested
	if helpFlag || hFlag {
		showHelp()
		return nil
	}

	// Load platforms from JSON file
	platforms, err := loadPlatforms(configFlag)
	if err != nil {
		return fmt.Errorf("加载平台时出错: %v", err)
	}

	// List platforms if requested
	if listFlag {
		listPlatforms(platforms)
		return nil
	}

	// Delete platform if requested
	if deleteFlag != "" {
		platforms, err = deletePlatform(platforms, deleteFlag)
		if err != nil {
			return fmt.Errorf("删除平台时出错: %v", err)
		}
		err = savePlatforms(platforms, configFlag)
		if err != nil {
			return fmt.Errorf("保存平台时出错: %v", err)
		}
		color.Green("平台 '%s' 删除成功!\n", deleteFlag)
		return nil
	}

	// Add platform if requested
	if addFlag {
		newPlatform, err := addNewPlatform()
		if err != nil {
			return fmt.Errorf("添加平台时出错: %v", err)
		}
		
		// 检查是否已存在同名平台，如果存在则更新而不是添加
		existingIndex := -1
		for i, platform := range platforms {
			if platform.Name == newPlatform.Name {
				existingIndex = i
				break
			}
		}
		
		if existingIndex != -1 {
			// 更新现有平台
			platforms[existingIndex] = newPlatform
			color.Yellow("平台 '%s' 已存在，正在更新配置...\n", newPlatform.Name)
		} else {
			// 添加新平台
			platforms = append(platforms, newPlatform)
		}
		
		err = savePlatforms(platforms, configFlag)
		if err != nil {
			return fmt.Errorf("保存平台时出错: %v", err)
		}
		color.Green("平台 '%s' 保存成功!\n", newPlatform.Name)
		return nil
	}

	// Launch Claude Code with selected platform
	err = launchClaudeWithPlatform(platforms, platformFlag, configFlag)
	if err != nil {
		return fmt.Errorf("启动 Claude Code 时出错: %v", err)
	}

	return nil
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
		color.Green("使用平台: %s\n", selectedPlatform.Name)
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

	// Loop until user confirms launch or chooses to exit
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

	// Set environment variables
	setEnvironment(selectedPlatform)
	color.Green("环境变量设置成功!")

	// Launch Claude Code
	color.Cyan("正在启动 Claude Code...")
	launchClaudeCode()
	return nil
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
		os.MkdirAll(dir, 0755)
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	color.Cyan("执行命令: claude")
	err := cmd.Run()
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

	color.Green("Claude Code 已成功启动!")
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

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败 %s: %v", path, err)
	}

	color.Green("配置已保存到: %s", path)
	return nil
}

// showHelp displays the help message
func showHelp() {
	color.Cyan("Claude Code 平台选择器 v1.1.0")
	color.Cyan("==============================")
	fmt.Println("一个用于选择和启动不同平台配置的 Claude Code 工具。")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  cctool [选项]")
	fmt.Println()
	color.Yellow("选项:")
	fmt.Println("  -list          列出所有可用平台")
	fmt.Println("  -platform name 直接使用指定平台")
	fmt.Println("  -f path        指定配置文件路径")
	fmt.Println("  -add           添加新平台")
	fmt.Println("  -delete name   按名称删除平台")
	fmt.Println("  -help, -h      显示此帮助信息")
	fmt.Println("  -version, -v   显示版本信息")
	fmt.Println()
	fmt.Println("配置文件格式:")
	fmt.Println("  {")
	fmt.Println("    \"platforms\": [")
	fmt.Println("      {")
	fmt.Println("        \"name\": \"平台名称\",")
	fmt.Println("        \"vendor\": \"厂商名称（可选）\",")
	fmt.Println("        \"ANTHROPIC_BASE_URL\": \"API基础URL\",")
	fmt.Println("        \"ANTHROPIC_AUTH_TOKEN\": \"认证令牌\",")
	fmt.Println("        \"ANTHROPIC_MODEL\": \"模型名称\",")
	fmt.Println("        \"ANTHROPIC_SMALL_FAST_MODEL\": \"快速小模型名称\"")
	fmt.Println("      }")
	fmt.Println("    ]")
	fmt.Println("  }")
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
