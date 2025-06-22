// cmd/cli/main.go
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/synapse/internal/agent"
	"github.com/synapse/internal/config"
	"github.com/synapse/internal/llm"
	"github.com/synapse/internal/llm/deepseek"
	"github.com/synapse/internal/llm/openai"
	"github.com/synapse/internal/tool"
	"github.com/synapse/internal/ui"

	"log"
	"os"
	"strings"
)

func main() {
	configPath := flag.String("config", "", "Path to the configuration file")

	flag.Parse()
	finalConfigPath, err := findConfigPath(*configPath)
	if err != nil {
		log.Fatalf("Could not find or create configuration: %v", err)
	}
	log.Printf("Using configuration file: %s", finalConfigPath)
	cfg, err := config.Load(finalConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	provider, err := createProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}
	log.Printf("Using LLM provider: %s", provider.Name())

	coreAgent := agent.New(provider)

	runCLI(coreAgent)
}

func createProvider(cfg *config.Config) (llm.LLMProvider, error) {
	providerConfig, ok := cfg.Providers[cfg.ActiveProvider]
	if !ok {
		return nil, fmt.Errorf("active provider '%s' not found in config", cfg.ActiveProvider)
	}

	switch providerConfig.Name {
	case "deepseek":
		return deepseek.New(providerConfig)
	case "openai":
		return openai.New(providerConfig)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerConfig.Name)
	}
}

func runCLI(coreAgent *agent.Agent) {
	ui.PrintWelcomeMessage()
	scanner := bufio.NewScanner(os.Stdin)

	for {

		fmt.Printf("%s ", ui.Blue("🔵 You>"))
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())

		if len(userInput) == 0 {
			continue
		}

		lowerInput := strings.ToLower(userInput)
		if lowerInput == "exit" || lowerInput == "quit" {
			fmt.Println(ui.BrightCyan("👋 Goodbye!"))
			break
		}
		if handled := handleLocalCommand(userInput, coreAgent); handled {
			continue
		}

		// 在调用 agent 之前，启动加载动画
		ui.StartSpinner("Thinking...")

		ctx := context.Background()
		responseChan, err := coreAgent.ProcessUserMessage(ctx, userInput)
		if err != nil {
			//  如果 agent 立即返回错误，也要停止动画
			ui.StopSpinner()
			log.Printf(ui.Red("Error processing message: %v"), err)
			continue
		}

		// 我们需要一个变量来跟踪是否已经打印了助手的头部信息
		assistantPrefixPrinted := false

		//  在循环从 channel 读取数据之前，停止动画
		// 但是，我们需要确保在打印任何内容之前停止它。
		// 最好的时机是在我们收到第一个 token 之后。
		for token := range responseChan {
			if !assistantPrefixPrinted {
				// 这是我们收到的第一个 token
				// 在打印它之前，停止动画并打印助手的前缀
				ui.StopSpinner()
				ui.PrintAssistantPrefix()
				assistantPrefixPrinted = true
			}
			// 打印收到的 token
			fmt.Print(ui.Green(token))
		}

		// 确保即使 channel 为空（例如只有工具调用，没有文本输出），动画也能被停止
		if !assistantPrefixPrinted {
			ui.StopSpinner()
		}

		fmt.Println() // 在每次对话结束后换行
	}
}

func handleLocalCommand(input string, coreAgent *agent.Agent) bool {
	parts := strings.Fields(input)
	command := strings.ToLower(parts[0])

	switch command {
	case "/add":
		if len(parts) < 2 {
			fmt.Println(ui.Red("Usage: /add <path/to/file>"))
			return true
		}
		pathToAdd := parts[1]
		content, err := os.ReadFile(pathToAdd)
		if err != nil {
			fmt.Printf(ui.Red("Error reading file '%s': %v\n"), pathToAdd, err)
			return true
		}

		coreAgent.AddFileToContext(pathToAdd, string(content))
		fmt.Printf(ui.Green("✓ File '%s' added to context. You can now ask questions about it.\n"), pathToAdd)

	case "/reset":
		coreAgent.ResetSession()
		fmt.Println(ui.Green("✓ Conversation has been reset."))
		return true

	case "/tools":
		fmt.Println(ui.Blue("--- Available Tools ---"))
		for _, t := range tool.GetDefaultTools() {
			fmt.Printf("  %s %s: %s\n", ui.BrightCyan("•"), ui.Cyan(t.Function.Name), t.Function.Description)
		}
		fmt.Println(ui.Blue("-----------------------"))
		return true
	case "/exit":
		fmt.Println(ui.BrightCyan("👋 Goodbye!"))
		os.Exit(0)

	default:
		fmt.Printf(ui.Yellow("Unknown command: %s. Type /help for a list of commands.\n"), command)
	}

	return false
}

// findConfigPath 按优先级搜索配置文件路径
func findConfigPath(explicitPath string) (string, error) {
	// 1. 命令行标志具有最高优先级
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", fmt.Errorf("config file specified by --config flag not found at %s", explicitPath)
	}

	// 2. 搜索用户配置目录
	currentUser, err := user.Current()
	if err == nil {
		userConfigPath := filepath.Join(currentUser.HomeDir, ".synapse", "config", "config.yaml")
		if _, err := os.Stat(userConfigPath); err == nil {
			return userConfigPath, nil
		}
	}

	// 3. 搜索当前工作目录
	currentDirConfigPath := "config.yaml"
	if _, err := os.Stat(currentDirConfigPath); err == nil {
		return currentDirConfigPath, nil
	}

	// 4. 如果都找不到，则尝试自动创建默认配置
	if currentUser != nil {
		defaultConfigDir := filepath.Join(currentUser.HomeDir, ".synapse", "config")
		defaultConfigPath := filepath.Join(defaultConfigDir, "config.yaml")

		// 询问用户是否创建
		fmt.Printf("Configuration file not found. Create a default one at %s? [Y/n]: ", defaultConfigPath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(response)) == "y" || response == "\n" {
			return createDefaultConfig(defaultConfigDir, defaultConfigPath)
		}
	}

	return "", errors.New("no configuration file found")
}

// createDefaultConfig 创建一个默认的配置文件
func createDefaultConfig(dir, path string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// 默认配置内容
	defaultContent := `
# Default configuration for Synapse
active_provider: "deepseek" # or "openai"

providers:
  deepseek:
    name: "deepseek"
    api_key_env: "DEEPSEEK_API_KEY"
    base_url: "https://api.deepseek.com"
    default_model: "deepseek-reasoner"
  openai:
    name: "openai"
    api_key_env: "OPENAI_API_KEY"
    base_url: "https://api.openai.com/v1"
    default_model: "gpt-4-turbo"
`
	err := os.WriteFile(path, []byte(strings.TrimSpace(defaultContent)), 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s\n", ui.Green("✓ Default config created. Please edit it and add your API keys to your environment variables (e.g., in ~/.bashrc or ~/.zshrc)."))

	// 首次创建后，可以直接退出，让用户去配置密钥
	os.Exit(0)
	return path, nil
}
