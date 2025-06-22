// internal/agent/agent.go
package agent

import (
	"context"

	"github.com/synapse/internal/llm"
)

type Agent struct {
	llmProvider llm.LLMProvider
	session     *Session
}

func New(provider llm.LLMProvider) *Agent {
	return &Agent{
		llmProvider: provider,
		session:     NewSession(),
	}
}

func (a *Agent) ProcessUserMessage(ctx context.Context, userInput string) (<-chan string, error) {
	a.session.AddUserMessage(userInput)

	outputChan := make(chan string)

	go a.handleStreaming(ctx, outputChan)

	return outputChan, nil
}

// AddFileToContext 是一个给 CLI 用的辅助方法
func (a *Agent) AddFileToContext(path, content string) {
	a.session.AddFileToContext(path, content)
}

func (a *Agent) ResetSession() {
	a.session.Reset()
}
