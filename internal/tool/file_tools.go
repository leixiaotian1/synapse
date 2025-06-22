// internal/tool/file_tools.go
package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/synapse/internal/llm"

	openai "github.com/sashabaranov/go-openai"
)

// registerFileTools 向注册表注册所有文件操作工具
func registerFileTools() {
	Register(
		llm.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "read_file",
				Description: "Read the content of a single file from the filesystem",
				Parameters: json.RawMessage(`{
					"type": "object", "properties": {"file_path": {"type": "string", "description": "The path to the file to read"}}, "required": ["file_path"]
				}`),
			},
		},
		toolReadFile,
	)

	Register(
		llm.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_file",
				Description: "Create a new file or overwrite an existing file with the provided content",
				Parameters: json.RawMessage(`{
					"type": "object", "properties": {"file_path": {"type": "string", "description": "The path where the file should be created"}, "content": {"type": "string", "description": "The content to write to the file"}}, "required": ["file_path", "content"]
				}`),
			},
		},
		toolCreateFile,
	)

	Register(
		llm.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "edit_file",
				Description: "Edit an existing file by replacing a specific snippet with new content",
				Parameters: json.RawMessage(`{
					"type": "object", "properties": {"file_path": {"type": "string", "description": "The path to the file to edit"}, "original_snippet": {"type": "string", "description": "The exact text snippet to find and replace"}, "new_snippet": {"type": "string", "description": "The new text to replace the original snippet with"}}, "required": ["file_path", "original_snippet", "new_snippet"]
				}`),
			},
		},
		toolEditFile,
	)
}

// --- Tool Implementations ---

func toolReadFile(arguments string) (string, error) {
	var args struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for read_file: %w", err)
	}
	content, err := readLocalFile(args.FilePath)
	if err != nil {
		return "", fmt.Errorf("error reading file '%s': %w", args.FilePath, err)
	}
	return fmt.Sprintf("Content of file '%s':\n\n%s", args.FilePath, content), nil
}

func toolCreateFile(arguments string) (string, error) {
	var args struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for create_file: %w", err)
	}
	if err := createFile(args.FilePath, args.Content); err != nil {
		return "", fmt.Errorf("error creating file '%s': %w", args.FilePath, err)
	}
	return fmt.Sprintf("Successfully created/updated file '%s'", args.FilePath), nil
}

func toolEditFile(arguments string) (string, error) {
	var args struct {
		FilePath        string `json:"file_path"`
		OriginalSnippet string `json:"original_snippet"`
		NewSnippet      string `json:"new_snippet"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for edit_file: %w", err)
	}

	content, err := readLocalFile(args.FilePath)
	if err != nil {
		return "", err
	}

	occurrences := strings.Count(content, args.OriginalSnippet)
	if occurrences == 0 {
		return "", errors.New("original snippet not found in file")
	}
	if occurrences > 1 {
		return "", fmt.Errorf("ambiguous edit: %d matches found for the snippet", occurrences)
	}

	updatedContent := strings.Replace(content, args.OriginalSnippet, args.NewSnippet, 1)
	if err := createFile(args.FilePath, updatedContent); err != nil {
		return "", err
	}
	return fmt.Sprintf("Successfully applied edit to file '%s'", args.FilePath), nil
}

// --- Helper functions ---

func normalizePath(p string) (string, error) {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("invalid path: contains '..'")
	}
	return absPath, nil
}

func readLocalFile(path string) (string, error) {
	normalizedPath, err := normalizePath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(normalizedPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func createFile(path, content string) error {
	normalizedPath, err := normalizePath(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(normalizedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(normalizedPath, []byte(content), 0644)
}
