// internal/llm/types.go
package llm

import openai "github.com/sashabaranov/go-openai"

// 这些是我们与 LLM API 交互时使用的核心数据结构
// 使用别名可以方便未来切换底层库而无需大规模重构
type Message = openai.ChatCompletionMessage
type Tool = openai.Tool
type ToolCall = openai.ToolCall
type ChatCompletionRequest = openai.ChatCompletionRequest
type ChatCompletionStream = openai.ChatCompletionStream
