package tool

import "fmt"

// Execute 是执行一个已注册工具的通用入口。
// 它接收工具名称和 JSON 格式的参数字符串。
// 它返回工具执行后的字符串结果或一个错误。
func Execute(name, arguments string) (string, error) {
	// 从注册表中查找对应的工具执行函数
	executorFunc, found := GetExecutor(name)
	if !found {
		return "", fmt.Errorf("tool '%s' not found in registry", name)
	}

	// 调用找到的函数并返回其结果
	return executorFunc(arguments)
}
