// internal/llm/deepseek/deepseek.go
package deepseek

import (
	"context"
	"fmt"

	"github.com/synapse/internal/config"
	"github.com/synapse/internal/llm"
	"github.com/synapse/internal/tool"

	openai "github.com/sashabaranov/go-openai"
)

type DeepSeekProvider struct {
	client *openai.Client
	config config.ProviderConfig
}

func New(cfg config.ProviderConfig) (llm.LLMProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key for deepseek is not set (env var: %s)", cfg.APIKeyEnv)
	}
	oaiConfig := openai.DefaultConfig(cfg.APIKey)
	oaiConfig.BaseURL = cfg.BaseURL
	client := openai.NewClientWithConfig(oaiConfig)

	return &DeepSeekProvider{
		client: client,
		config: cfg,
	}, nil
}

func (p *DeepSeekProvider) Name() string {
	return p.config.Name
}

func (p *DeepSeekProvider) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest) (*llm.ChatCompletionStream, error) {
	if req.Model == "" {
		req.Model = p.config.DefaultModel
	}
	return p.client.CreateChatCompletionStream(ctx, req)
}

func (p *DeepSeekProvider) GetTools() []llm.Tool {
	return tool.GetDefaultTools()
}
