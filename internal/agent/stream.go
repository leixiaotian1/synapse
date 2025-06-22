// internal/agent/stream.go
package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/synapse/internal/llm"
	"github.com/synapse/internal/tool"
	"github.com/synapse/internal/ui"

	"github.com/sashabaranov/go-openai" // 需要这个来获取 openai.ToolTypeFunction
)

// handleStreaming 是 agent 的核心循环
func (a *Agent) handleStreaming(ctx context.Context, outputChan chan<- string) {
	defer close(outputChan)

	// 循环直到获得最终的文本响应，或者发生不可恢复的错误
	// 添加一个循环次数限制，防止无限循环
	const maxTurns = 10
	for i := 0; i < maxTurns; i++ {
		a.session.TrimHistory()
		req := llm.ChatCompletionRequest{
			// Model 字段不在这里设置，让 provider 来决定默认值
			Messages: a.session.GetHistory(),
			Tools:    a.llmProvider.GetTools(),
			Stream:   true,
		}

		stream, err := a.llmProvider.CreateChatCompletionStream(ctx, req)
		if err != nil {
			// **修复 #1: 使用 fmt.Sprintf 然后发送到 channel**
			errorMsg := fmt.Sprintf("API Error: %v", err)
			outputChan <- ui.Red(errorMsg) // 使用 UI 包来给错误上色
			return
		}

		var fullResponse strings.Builder
		var accumulatedToolCalls []llm.ToolCall

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				// **修复 #2: 使用 fmt.Sprintf 然后发送到 channel**
				errorMsg := fmt.Sprintf("Stream Error: %v", err)
				outputChan <- ui.Red(errorMsg)
				stream.Close()
				return
			}

			delta := response.Choices[0].Delta
			if delta.Content != "" {
				fullResponse.WriteString(delta.Content)
				// 将正常的 token 直接发送出去
				outputChan <- delta.Content
			}
			if len(delta.ToolCalls) > 0 {
				accumulatedToolCalls = accumulateToolCalls(accumulatedToolCalls, delta.ToolCalls)
			}
		}
		stream.Close()

		// 记录助手的回复，即使是空的，也要记录工具调用
		a.session.AddAssistantMessage(fullResponse.String(), accumulatedToolCalls)

		if len(accumulatedToolCalls) > 0 {
			// 如果有工具调用，执行它们并继续循环
			executeToolCalls(a.session, accumulatedToolCalls, outputChan)
			continue
		}

		// 没有工具调用，这是对话的终点，循环结束
		return
	}

	// 如果循环达到最大次数，发送一个警告信息
	outputChan <- ui.Yellow("\nWarning: Maximum conversation turns reached.")
}

// accumulateToolCalls 从流式响应中逐步构建完整的工具调用列表
func accumulateToolCalls(existing, delta []llm.ToolCall) []llm.ToolCall {
	for _, tcDelta := range delta {
		if tcDelta.Index == nil {
			continue
		}
		idx := *tcDelta.Index
		for len(existing) <= idx {
			existing = append(existing, llm.ToolCall{})
		}
		if tcDelta.ID != "" {
			existing[idx].ID = tcDelta.ID
		}
		if tcDelta.Type != "" {
			existing[idx].Type = openai.ToolTypeFunction
		}
		if tcDelta.Function.Name != "" {
			existing[idx].Function.Name += tcDelta.Function.Name
		}
		if tcDelta.Function.Arguments != "" {
			existing[idx].Function.Arguments += tcDelta.Function.Arguments
		}
	}
	return existing
}

// executeToolCalls 执行工具并将结果添加到会话中
// **新增 outputChan 参数，用于在工具执行时提供反馈**
func executeToolCalls(session *Session, toolCalls []llm.ToolCall, outputChan chan<- string) {
	outputChan <- fmt.Sprintf("\n%s", ui.BrightCyan(fmt.Sprintf("⚡ Executing %d function call(s)...", len(toolCalls))))

	for _, tc := range toolCalls {
		// 为了更好的用户体验，将工具调用的信息也发送到 UI
		outputChan <- fmt.Sprintf("\n%s %s", ui.Blue("→ Calling tool:"), ui.Cyan(tc.Function.Name))

		result, err := tool.Execute(tc.Function.Name, tc.Function.Arguments)

		if err != nil {
			errorMsg := fmt.Sprintf("Error executing tool '%s': %v", tc.Function.Name, err)
			// 将错误信息发送到 UI
			outputChan <- fmt.Sprintf("\n%s", ui.Red(errorMsg))
			// 并记录到会话历史中
			session.AddToolMessage(tc.ID, errorMsg)
		} else {
			// 工具成功执行，记录结果
			session.AddToolMessage(tc.ID, result)
			// 也可以选择性地将成功信息发送到 UI
			outputChan <- fmt.Sprintf("\n%s", ui.Green("✓ Tool finished."))
		}
	}
	outputChan <- fmt.Sprintf("\n%s", ui.BrightCyan("🔄 Resuming conversation..."))
}
