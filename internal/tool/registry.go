// internal/tool/registry.go
package tool

import (
	"fmt"

	"github.com/synapse/internal/llm"

	"sync"
)

var (
	registry     = make(map[string]ToolFunc)
	definitions  []llm.Tool // <-- 直接使用 llm.Tool 类型
	registryOnce sync.Once
)

// Register 注册一个工具，使其可用。
// 它接收一个 llm.Tool 定义和一个我们自定义的 ToolFunc 实现。
func Register(def llm.Tool, fn ToolFunc) { // <-- 参数类型改为 llm.Tool
	name := def.Function.Name
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("tool %s is already registered", name))
	}
	registry[name] = fn
	definitions = append(definitions, def)
}

// GetExecutor 返回一个可以执行工具的函数。
func GetExecutor(name string) (ToolFunc, bool) {
	fn, found := registry[name]
	return fn, found
}

// GetDefaultTools 返回所有已注册工具的定义。
func GetDefaultTools() []llm.Tool { // <-- 返回类型改为 []llm.Tool
	InitTools() // 确保工具已初始化
	return definitions
}

// InitTools 初始化并注册所有默认工具。
func InitTools() {
	registryOnce.Do(func() {
		registerFileTools()
	})
}
