# Synapse 🧠 - Your AI Coding Familiar
[![Go Version](https://img.shields.io/github/go-mod/go-version/github.com/synapse?style=for-the-badge)](https://golang.org/)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=for-the-badge)](CONTRIBUTING.md)

**Synapse** is a highly extensible command-line AI coding assistant built in Go. It's more than just a code generator — it's an "AI coding familiar" that understands your intent, securely interacts with the local file system, and collaborates with you like a real pair-programming partner.

By combining the reasoning power of large language models (LLMs) such as DeepSeek and OpenAI GPT with local tool execution, Synapse seamlessly transforms your ideas into high-quality code.

---

## ✨ Features

- **🧠 Intelligent Dialogue & Reasoning:** Powered by cutting-edge LLMs, capable of understanding complex coding needs, analyzing existing code, and formulating execution plans.
- **🔌 Pluggable Model Support:** Easily switch between different LLM providers (e.g., DeepSeek, OpenAI) via simple configuration.
- **🛠️ Secure File System Interaction:** Interact safely with local files through a strictly defined set of tools (function calls).
- **🤖 Proactive Workflow:** Synapse announces its plan, reports results, and suggests next steps, offering a fully transparent workflow.
- **🚀 Optimized User Experience:**
  - Beautifully colored terminal interface with clear and readable information.
  - Loading animations while waiting for AI responses, eliminating perceived lag.
  - Streaming responses let you see the AI’s thought process and code generation in real-time.
- **⚙️ Extensible Architecture:** Modular design makes it easy to add new LLM providers, tools, or interfaces (e.g., mcp, Web UI).
- **🔧 Flexible Configuration:** Manage all settings via `config.yaml` and `.env` files rather than hardcoding them.

---

## 🚀 Quick Start

### 1. Prerequisites

- [Go](https://go.dev/doc/install) (version 1.24.1 or higher)
- API keys from one or more LLM providers (e.g., DeepSeek, OpenAI)

### 2. Installation

To build and run Synapse:

```bash
# Build the binary
go build -o synapse ./cmd/cli/main.go

# Run the CLI
./synapse
```

Or directly run without building:

```bash
go run ./cmd/cli/main.go
```

### 3. Configuration

when you run Synapse for the first time, it will create a `config.json` file in your {HomeDir}/.synapse/config and will prompt you to configure your preferred provider.


####  Modify `config.yaml`

Update the `config.yaml` file to configure your preferred provider:

```yaml
# config.yaml

# ... (provider definitions)

# Switch your active provider here: "deepseek" or "openai"
active_provider: "deepseek"
```
now you can run Synapse in your terminal.
---


---

## 🤝 Contributing

Contributions are always welcome! Whether you're fixing bugs, improving performance, adding features, or enhancing documentation, feel free to submit a PR.

---

## 📄 License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.
