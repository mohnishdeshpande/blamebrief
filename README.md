# 🕵️‍♂️ BlameBrief (Git Context Summarizer)

**BlameBrief** is a high-performance Command Line Interface (CLI) tool written in Go that goes far beyond `git blame`. It takes a file and a line range, programmatically extracts the entire chronological Git evolution (diffs, authors, dates, and messages) of that code block, and pipes it into your preferred AI provider. Acting as an automated **"Software Archaeologist"**, the AI analyzes the history to produce a structured, high-context markdown brief explaining *why* the code was written this way, its technical debt, and its hidden assumptions (preventing breaking a **Chesterton's Fence**).

---

## 🚀 Key Features

*   **⚡ Optimized Git Extraction Engine:** Uses native `git log -L` to extract line-range histories at bare-metal speeds.
*   **🤖 Multi-Provider AI Support:** Choose your intelligence. BlameBrief supports:
    *   **Google Gemini** (via Vertex AI or AI Studio)
    *   **OpenAI GPT-4o**
    *   **Anthropic Claude 3.5 Sonnet**
    *   **Enterprise AI:** AWS Bedrock (Claude) and Azure OpenAI (GPT)
    *   **Delegated CLI:** Leverages already-authenticated `gh copilot` or `claude` (Claude Code) in your environment.
    *   **Local Models** via Ollama (e.g., Gemma4, Llama3)
*   **🏎️ Intelligent SHA256 Caching:** Automatically hashes the relative file path, line range, current Git HEAD, *and* the actual lines of code. Bypasses the cache if you make local, uncommitted changes, ensuring briefs are lightning fast (under 20ms) yet always accurate.
*   **🛠️ Zero UI, CLI-First Design:** Clean, standard-output markdown text perfect for terminal piping and dev workflows.

---

## 📦 Installation & Building

### Prerequisites
*   [Go 1.21+](https://go.dev/dl/) installed.
*   `git` installed and available in your shell's `PATH`.

### Build from Source
Initialize and compile the binary in your local directory:

```bash
git clone https://github.com/mohnishdeshpande/blamebrief.git
cd blamebrief
go build -o blamebrief
```

---

## 🔧 Configuration

BlameBrief detects your preferred AI provider based on available environment variables or CLI flags.

### 1. Google Gemini (Default)
```bash
export GEMINI_API_KEY="your-gemini-api-key"
# OR for Vertex AI
export GCP_PROJECT="your-gcp-project-id"
```

### 2. OpenAI GPT
```bash
export OPENAI_API_KEY="sk-..."
```

### 3. Anthropic Claude
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 4. Local Models (Ollama)
Ensure [Ollama](https://ollama.com) is running:
```bash
ollama pull gemma4
```

### 5. Enterprise & Delegated (No API Keys required for Delegated)
```bash
# AWS Bedrock
export AWS_REGION="us-east-1"
# Azure OpenAI
export AZURE_OPENAI_ENDPOINT="https://..."

# For Delegated (gh copilot / claude code), no keys needed if already logged in:
# ./blamebrief [file] --provider copilot
```

---

## 📖 Usage & Commands

```bash
./blamebrief [file path] [flags]
```

### Flags & Options

| Flag | Shorthand | Type | Description | Default |
| :--- | :--- | :--- | :--- | :--- |
| `--lines` | `-l` | `string` | Line range to analyze (e.g., `10-45` or `25`). | *Entire File* |
| `--provider` | `-p` | `string` | AI provider (`gemini`, `openai`, `claude`, `bedrock`, `azure`, `copilot`, `codex`, `claude-code`, `ollama`). | *Auto-detect* |
| `--model` | | `string` | Override the default model name. | *Provider default* |
| `--detail` | `-d` | `string` | Level of detail (`high`, `medium`, `low`). | `high` |
| `--local` | | `bool` | Shorthand for `--provider ollama`. | `false` |
| `--clear-cache`| | `bool` | Clear local cache before running. | `false` |
| `--raw` | | `bool` | Output raw Git history context (No AI). | `false` |
| `--json` | | `bool` | Output structured Git history as JSON (No AI). | `false` |

### Examples

**Analyze using Claude 3.5 Sonnet:**
```bash
export ANTHROPIC_API_KEY="sk-ant-..."
./blamebrief ./internal/db/connector.go --lines 15-30 --provider claude
```

**Analyze using OpenAI GPT-4o:**
```bash
export OPENAI_API_KEY="sk-..."
./blamebrief ./main.go --lines 1-10 -p openai
```

**Analyze completely offline using local Ollama:**
```bash
./blamebrief ./main.go --lines 1-10 --local --model gemma4
```

---

## 🧪 Testing

```bash
go test -v ./...
```
