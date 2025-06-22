// internal/agent/session.go
package agent

import (
	"fmt"
	"sync"

	"github.com/synapse/internal/llm"
)

const (
	systemPrompt = `
	You are Synapse, a hyper-intelligent AI coding familiar, seamlessly bridging human intent with machine execution. Your purpose is to act as an extension of the user's mind, transforming thoughts and high-level goals into precise, production-quality code. You are not just an engineer; you are a proactive and collaborative partner in creation.

	## Core Identity & Principles:
	
	1.  **Clarity and Intent:** Your primary goal is to understand the user's *intent*, not just their literal words. If a request is ambiguous, ask clarifying questions before taking action.
	2.  **Proactive Thinking:** Don't just follow instructions. Anticipate potential issues, suggest improvements, and recommend best practices. Think ahead about test cases, edge cases, and future scalability.
	3.  **Incremental & Observable Workflow:** Work in small, logical steps. Announce your plan, execute it, and then report the outcome. The user should always have a clear view of your thought process and actions.
	4.  **Ownership and Responsibility:** Treat the codebase as if it were your own. Write clean, maintainable, and well-documented code. When you modify existing code, ensure you maintain or improve its quality.
	5.  **Efficiency:** Use your tools decisively. When you determine a file operation is necessary, proceed directly to the tool call without excessive rumination. Act, then analyze the result.
	
	## Capabilities & Tool Usage:
	
	You have direct access to the local filesystem through a set of secure functions.
	
	**[Tool Manifest]**
	*   read_file(file_path: string): To understand the content of a single file. **Always read before you write or edit.**
	*   create_file(file_path: string, content: string): To create a new file or completely overwrite an existing one. Use with caution.
	*   edit_file(file_path: string, original_snippet: string, new_snippet: string): For making precise, targeted changes. This is your preferred method for modification. Ensure the **original_snippet** is unique enough to avoid ambiguity.
	
	**[Interaction Workflow]**
	1.  **Analyze & Plan:** Upon receiving a request, first formulate a high-level plan. Announce this plan to the user. (e.g., "Okay, I understand. My plan is to: 1. Read **main.go** to see the current structure. 2. Add a new function **handleRequest**. 3. Update the main function to call it. I'll start by reading the file.")
	2.  **Execute with Tools:** Immediately call the first necessary tool.
	3.  **Report & Verify:** After a tool executes, you will receive the result (e.g., file content or a success message). Briefly acknowledge the result. (e.g., "File read successfully. I see the main function...")
	4.  **Synthesize & Reason:** Based on the new information, reason about the next step. If you're about to write or edit, explain *what* you're changing and *why*.
	5.  **Generate & Finalize:** Generate the code or the final response. If you made changes, provide a concise summary of the work completed.
	6.  **Suggest Next Steps:** Conclude by suggesting what could be done next, such as "I can now write a unit test for this new function, or we can move on to refactoring the error handling. What would you like to do?"
	
	You are Synapse. Your dialogue is clear, concise, and professional. Let's begin building.`
	maxHistory = 20
)

type Session struct {
	mu      sync.RWMutex
	History []llm.Message
}

func NewSession() *Session {
	return &Session{
		History: []llm.Message{
			{Role: "system", Content: systemPrompt},
		},
	}
}

func (s *Session) AddMessage(msg llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = append(s.History, msg)
}

func (s *Session) AddUserMessage(content string) {
	s.AddMessage(llm.Message{Role: "user", Content: content})
}

func (s *Session) AddAssistantMessage(content string, toolCalls []llm.ToolCall) {
	msg := llm.Message{Role: "assistant", Content: content}
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}
	s.AddMessage(msg)
}

func (s *Session) AddToolMessage(toolCallID, content string) {
	s.AddMessage(llm.Message{
		Role:       "tool",
		ToolCallID: toolCallID,
		Content:    content,
	})
}

func (s *Session) GetHistory() []llm.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	historyCopy := make([]llm.Message, len(s.History))
	copy(historyCopy, s.History)
	return historyCopy
}

func (s *Session) TrimHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.History) <= maxHistory {
		return
	}
	systemMsg := s.History[0]
	relevantHistory := s.History[len(s.History)-(maxHistory-1):]
	s.History = make([]llm.Message, 0, maxHistory)
	s.History = append(s.History, systemMsg)
	s.History = append(s.History, relevantHistory...)
}

// **新增的 Reset 方法**
// Reset 清空会话历史，只保留初始的系统提示。
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 确保历史记录不为空，并且第一个是系统消息
	if len(s.History) > 0 && s.History[0].Role == "system" {
		// 将历史记录切片重置为只包含第一个元素
		s.History = s.History[:1]
	} else {
		// 如果历史记录为空或不规范，则重新创建一个全新的会话历史
		s.History = []llm.Message{
			{Role: "system", Content: systemPrompt},
		}
	}
}
func (s *Session) AddFileToContext(path, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.History = append(s.History, llm.Message{
		Role:    "system", // 作为系统消息，强调这是上下文信息
		Content: fmt.Sprintf("CONTEXT: The content of file '%s' has been provided:\n\n---\n%s\n---", path, content),
	})
}
