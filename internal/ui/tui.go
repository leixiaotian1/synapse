// internal/ui/tui.go
package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var (
	Blue       = color.New(color.FgBlue, color.Bold).SprintFunc()
	Cyan       = color.New(color.FgCyan).SprintFunc()
	BrightCyan = color.New(color.FgHiCyan, color.Bold).SprintFunc()
	Green      = color.New(color.FgGreen).SprintFunc()
	Red        = color.New(color.FgRed).SprintFunc()
	Yellow     = color.New(color.FgYellow).SprintFunc()
	Dim        = color.New(color.Faint).SprintFunc()
)

// PrintWelcomeMessage 打印一个现代化、信息丰富的欢迎界面。
func PrintWelcomeMessage() {
	// 使用 BrightCyan 作为边框和标题颜色，使其更突出
	border := BrightCyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	title := `  Synapse Coding Assistant 🤖    `
	tagline := "Your AI Coding Familiar - Bridging thought and code."

	fmt.Println(border)
	fmt.Println(BrightCyan(title))
	fmt.Printf("%s\n", Dim(tagline))
	fmt.Println(border)
	fmt.Println()

	// 使用标题和分组来组织信息
	fmt.Printf("%s\n", Blue("✨ FEATURES:"))
	fmt.Printf("  %s %s\n", BrightCyan("•"), "Intelligent code generation & refactoring")
	fmt.Printf("  %s %s\n", BrightCyan("•"), "Interactive file system operations (read, create, edit)")
	fmt.Printf("  %s %s\n", BrightCyan("•"), "Pluggable LLM backend (OpenAI, DeepSeek, etc.)")
	fmt.Println()

	fmt.Printf("%s\n", Blue("🚀 HOW TO USE:"))
	fmt.Printf("  %s Just start chatting! Ask for code, refactoring, or ideas.\n", Cyan("1. General Chat:"))
	fmt.Printf("  %s Use the `/add` command to load a file's content into context.\n", Cyan("2. File Context:"))
	fmt.Printf("  %s Use the `/reset` command to create a new conversation to context.\n", Cyan("3. Reset Conversation:"))
	fmt.Printf("  %s Use the `/tools` command to show all tools.\n", Cyan("4. Show Tools:"))
	fmt.Printf("     %s %s\n", Dim("e.g."), Cyan("/add ./path/to/your/file.go"))
	fmt.Printf("  %s The AI can read and edit files by asking for permission.\n", Cyan("5. File Editing:"))
	fmt.Printf("     %s %s\n", Dim("e.g."), "Refactor the error handling in main.go")
	fmt.Println()

	fmt.Printf("%s\n", Blue("⚙️ COMMANDS & FLAGS:"))
	fmt.Printf("  %s or %s %s\n", Cyan("exit"), Cyan("quit"), Dim("- End the session."))
	fmt.Printf("  %s %s\n", Cyan("--config"), Dim("- Specify a path to your config file (e.g., --config my_config.yaml)."))
	fmt.Printf("  %s %s\n", Cyan("--help"), Dim("- Show all available command-line flags."))
	fmt.Println()

	fmt.Println(Dim("Ready to build. Type your request to begin..."))
	fmt.Println()
}

func PrintToolCall(name string) {
	fmt.Printf("\n%s %s\n", Blue("→ Calling tool:"), Cyan(name))
}

func PrintAssistantPrefix() {
	fmt.Printf("\n%s ", Green("🤖 Assistant>"))
}

func PrintReasoning(content string) {
	// Simple reasoning print for now
	fmt.Printf("%s", Dim(content))
}

var (
	// a channel to signal the spinner to stop
	stopSpinnerChan = make(chan struct{})
	spinnerChars    = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// StartSpinner 在后台启动一个旋转的加载动画。
// 它接收一个消息字符串（例如 "Thinking..."）显示在动画旁边。
func StartSpinner(message string) {
	// 确保之前的 channel 是关闭的，或者重新创建一个新的
	stopSpinnerChan = make(chan struct{})

	go func() {
		i := 0
		for {
			select {
			case <-stopSpinnerChan:
				// 接收到停止信号，清除当前行并退出 goroutine
				fmt.Print("\r\033[K") // 清除行的 ANSI escape code
				return
			default:
				// 打印动画的当前帧和消息
				// '\r' 将光标移到行首，这样下一帧就可以覆盖当前帧
				spinnerFrame := BrightCyan(spinnerChars[i%len(spinnerChars)])
				fmt.Printf("\r%s %s ", spinnerFrame, message)
				time.Sleep(100 * time.Millisecond) // 控制动画速度
				i++
			}
		}
	}()
}

// StopSpinner 发送信号以停止加载动画。
func StopSpinner() {
	// 使用 select 来防止在 channel 已经关闭时再次关闭它导致 panic
	select {
	case <-stopSpinnerChan:
		// Channel 已经关闭，什么都不做
		return
	default:
		// Channel 是打开的，关闭它来发送信号
		close(stopSpinnerChan)
	}
}
