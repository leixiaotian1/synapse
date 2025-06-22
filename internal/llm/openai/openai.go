// internal/llm/openai/openai.go
package openai

import (
	"context"
	"fmt"

	"github.com/synapse/internal/config"
	"github.com/synapse/internal/llm"
	"github.com/synapse/internal/tool"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider 实现了 llm.LLMProvider 接口，用于与 OpenAI API 交互。
type OpenAIProvider struct {
	client *openai.Client
	config config.ProviderConfig
}

// New 创建并返回一个新的 OpenAIProvider 实例。
func New(cfg config.ProviderConfig) (llm.LLMProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key for openai is not set (env var: %s)", cfg.APIKeyEnv)
	}

	// 对于标准的 OpenAI API，我们不需要设置 BaseURL，
	// go-openai 库会使用默认的 "https://api.openai.com/v1"。
	// 如果用户在 config.yaml 中提供了自定义的 base_url（例如用于代理），我们则使用它。
	oaiConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" && cfg.BaseURL != "https://api.openai.com/v1" {
		oaiConfig.BaseURL = cfg.BaseURL
	}

	client := openai.NewClientWithConfig(oaiConfig)

	return &OpenAIProvider{
		client: client,
		config: cfg,
	}, nil
}

// Name 返回提供商的名称。
func (p *OpenAIProvider) Name() string {
	return p.config.Name
}

// CreateChatCompletionStream 使用 OpenAI 客户端发起流式聊天请求。
func (p *OpenAIProvider) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest) (*llm.ChatCompletionStream, error) {
	// 如果请求中没有指定模型，则使用配置中的默认模型。
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}
	return p.client.CreateChatCompletionStream(ctx, req)
}

// GetTools 返回所有默认的可用工具。
// 因为我们的工具定义是通用的，所以 OpenAI 和 DeepSeek 可以共享同一套工具。
func (p *OpenAIProvider) GetTools() []llm.Tool {
	return tool.GetDefaultTools()
}
