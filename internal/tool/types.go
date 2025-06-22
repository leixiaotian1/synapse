package tool

// ToolFunc 是一个可执行工具的函数签名。
// 它接收由 LLM 生成的、JSON 格式的参数字符串，
// 并返回一个对 LLM 有意义的、字符串形式的结果，或者一个错误。
type ToolFunc func(arguments string) (string, error)
