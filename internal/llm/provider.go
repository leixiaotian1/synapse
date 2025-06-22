// internal/llm/provider.go
package llm

import (
	"context"
)

// LLMProvider 是所有 LLM 客户端必须实现的接口
type LLMProvider interface {
	// Name 返回提供商的名称 (e.g., "deepseek", "openai")
	Name() string
	// CreateChatCompletionStream 发起一个流式聊天请求
	CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionStream, error)
	// GetTools 返回该 provider 推荐使用的工具定义
	// 注意：工具定义是通用的，但某些模型可能对格式有特殊偏好
	GetTools() []Tool
}
