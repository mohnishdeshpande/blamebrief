# 🕵️‍♂️ BlameBrief (Git Context Summarizer)

**BlameBrief** is a high-performance Command Line Interface (CLI) tool written in Go that goes far beyond `git blame`. It takes a file and a line range, programmatically extracts the entire chronological Git evolution (diffs, authors, dates, and messages) of that code block, and pipes it into Gemini or local Gemma models. Acting as an automated **"Software Archaeologist"**, the AI analyzes the history to produce a structured, high-context markdown brief explaining *why* the code was written this way, its technical debt, and its hidden assumptions (preventing breaking a **Chesterton's Fence**).

---

## 🚀 Key Features

*   **⚡ Optimized Git Extraction Engine:** Uses native `git log -L` to extract line-range histories at bare-metal speeds.
*   **🧠 Gemini Long-Context Analysis:** Pipes massive commit history directly into Gemini, using the entire chronological evolution to tell the code's story. Supports both **Google AI Studio** and **Google Cloud Vertex AI** via the new unified GenAI Go SDK.
*   **🔒 Local & Private Fallback:** Support for running against local Gemma models via **Ollama** (e.g., `gemma4`), keeping sensitive enterprise code completely on your machine.
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
cd blamebrief-cli
go build -o blamebrief
```

---

## 🔧 Configuration

BlameBrief supports multiple backends depending on the environment variables defined:

### 1. Google AI Studio (Default)
Fastest and easiest setup for developers.
```bash
export GEMINI_API_KEY="your-gemini-api-key"
```

### 2. Vertex AI (Google Cloud)
Enterprise setup leveraging GCP authentication and Application Default Credentials (ADC).
```bash
export GCP_PROJECT="your-gcp-project-id"
export GCP_LOCATION="us-central1" # Optional, defaults to us-central1
```

### 3. Local Gemma Fallback (Ollama)
Completely offline, private execution using models running locally on your hardware.
Ensure [Ollama](https://ollama.com) is installed and running, then pull Gemma:
```bash
ollama pull gemma4:e2b # or gemma2
```

---

## 📖 Usage & Commands

```bash
./blamebrief [file path] [flags]
```

### Flags & Options

| Flag | Shorthand | Type | Description | Default |
| :--- | :--- | :--- | :--- | :--- |
| `--lines` | `-l` | `string` | Line range to analyze (e.g., `10-45` or `25`). If omitted, analyzes the whole file. | *Entire File* |
| `--detail` | `-d` | `string` | Level of detail for the AI report (`high`, `medium`, `low`). | `high` |
| `--local` | | `bool` | Execute offline using local Ollama instead of cloud Gemini. | `false` |
| `--model` | | `string` | Override the default model name (e.g. `gemini-3.1-pro`, `gemini-3.1-flash-lite`, `gemma4:e2b`). | `gemini-3.5-flash` / `gemma2` |
| `--clear-cache`| | `bool` | Clear local blamebrief cache files before running. | `false` |
| `--raw` | | `bool` | Output the raw gathered Git history context package as Markdown (Deterministic, No AI). | `false` |
| `--json` | | `bool` | Output the structured Git history context package as pretty-printed JSON (Deterministic, No AI). | `false` |

### Examples

**Analyze a specific line range using Google AI Studio (Gemini Flash):**
```bash
export GEMINI_API_KEY="AIzaSy..."
./blamebrief ./internal/db/connector.go --lines 15-30
```

**Analyze a file completely offline using a local Gemma 4 model:**
```bash
./blamebrief ./main.go --lines 1-10 --local --model gemma4:e2b
```

**Clear cache and run a fresh deep-dive analysis:**
```bash
./blamebrief ./pkg/git/extractor.go --clear-cache --detail high
```

**Deterministic Output Mode (Zero AI / Offline Data Collection):**
Analyze and export the compiled line-range history as clean Markdown or pretty-printed JSON instantly, without making any network calls, local model loads, or requiring AI credentials:
```bash
# Output the raw gathered Git history context package as Markdown
./blamebrief ./main.go --lines 1-10 --raw

# Output the structured Git history context package as pretty-printed JSON
./blamebrief ./main.go --lines 1-10 --json
```

---

## 📝 Example Output Report

When you run BlameBrief, it outputs an objective, highly competent markdown report structure:

```markdown
# BlameBrief: ./internal/db/connector.go (Lines 10-45)

**Executive Summary**: This block was originally a simple SQL connection. In 2022, it was wrapped in a retry loop to handle AWS RDS failovers. In 2023, the connection timeout was shortened because it was causing thread-pool exhaustion during peak traffic.

**Chronological Timeline**:
*   **Commit `a1b2c3d4` (Author: Sarah Lin, 2022-04-12)**: Initial SQL database connector implementation.
*   **Commit `e5f6g7h8` (Author: Alex Chen, 2022-09-18)**: Wrapped connection string in retry loop to mitigate AWS RDS failovers.
*   **Commit `i9j0k1l2` (Author: Sarah Lin, 2023-11-05)**: Shortened connection timeout to 2 seconds to prevent thread-pool exhaustion under heavy failover load.

**Architectural Intent & Chesterton's Fence Warning**:
*   **Hidden Purpose**: The connection pool's current `MaxIdleConns` setting is locked to 5 instead of Go's default. Removing this lock will likely re-introduce the 2023 thread exhaustion bug, where idle connections leaked during RDS master switches.
*   **Warning**: Do not remove the retry logic or expand the connection timeout without configuring client-side request limits first.

**Technical Debt & Sentiment Score**:
*   **Frustration/Sentiment**: Alex Chen's commit message was: "fix connection pool leaking AGAIN. RDS fails, we crash. Hotfix." implying a high level of stress during an active outage.
*   **Frustration Score**: 7/10
*   **Technical Debt**: The retry loop lacks progressive exponential backoff and sleeps for a static 500ms. This acts as a potential "thundering herd" if the database is down for a prolonged period.
```

---

## 🧪 Testing

We keep our technical execution production-grade. Run the full unit testing suite (which includes full mock servers for local AI integrations):

```bash
go test -v ./...
```
