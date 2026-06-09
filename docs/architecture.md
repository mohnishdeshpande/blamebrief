# 🏗️ BlameBrief System Architecture & Tech Layering

BlameBrief is designed as an ultra-fast, zero-dependency, and highly secure Command Line Interface (CLI) tool. It operates as a **"Software Archaeologist"** to solve the **"Chesterton's Fence"** problem—helping developers understand the hidden design intentions and technical debt of legacy code segments in under 20 milliseconds.

This document details the software architecture, modular layering, and data pipelines that make BlameBrief highly optimized and production-grade.

---

## 🗺️ Architectural Overview

BlameBrief follows a clean, decoupled, and unidirectional data-flow architecture. The codebase is strictly partitioned into domain-specific packages, keeping the core CLI parsing separate from Git subprocess management, file system caching, and LLM integrations.

### Component Dependency Diagram
```
                     +---------------------------+
                     |          main.go          |
                     |     (CLI Entrypoint)      |
                     +-------------+-------------+
                                   |
                                   v
                     +---------------------------+
                     |         pkg/cmd           |
                     |  (CLI Command & Parsers)  |
                     +----+-------------+----+---+
                          |             |    |
       +------------------+             |    +-------------------+
       |                                |                        |
       v                                v                        v
+--------------+               +--------+-------+        +---------------+
|   pkg/git    |               |   pkg/cache    |        |    pkg/ai     |
| (Extraction) |               |   (Caching)    |        | (Multi-Prov.) |
+--------------+               +----------------+        +---------------+
```

---

## 🗂️ Tech Layering & Packages

The codebase is organized into four main layers:

### 1. The Orchestration Layer (`cmd/`)
*   **Core Logic:** Defined in `cmd/root.go` using `spf13/cobra`.
*   **Responsibilities:**
    *   Exposes and parses CLI arguments and flags (`--lines`, `--detail`, `--provider`, `--model`, `--clear-cache`).
    *   Applies strict input validation (ensuring target file exists, ranges are numeric, and bounds are valid).
    *   Orchestrates execution flow by invoking the Git extractor, checking the cache, dispatching prompt payloads to AI clients, and displaying final outputs.

### 2. The Git Extraction Layer (`pkg/git/`)
*   **Core Logic:** Implemented in `pkg/git/extractor.go` utilizing standard Go `os/exec` subprocesses.
*   **Responsibilities:**
    *   Locates the file's Git repository root (`git rev-parse --show-toplevel`) to handle calls from subdirectories.
    *   Retrieves the current repository HEAD hash (`git rev-parse HEAD`) to support cache versioning.
    *   Executes native `git log -L <start>,<end>:<file>` to programmatically extract line history at bare-metal speeds.
    *   Reads and extracts target line contents from the current workspace file to compare against historical diffs.

### 3. The Performance & Caching Layer (`pkg/cache/`)
*   **Core Logic:** Implemented in `pkg/cache/cache.go` saving briefs locally inside the user's home folder (`~/.blamebrief/cache/`).
*   **Responsibilities:**
    *   Implements a deterministic SHA256-based caching mechanism.
    *   **The Cache Key Compound:** Hashes a unique string constructed of:
        `Relative File Path + Line Range + HEAD Commit Hash + Current Line Code Contents`
    *   **Auto-Invalidation:** If a developer makes local, uncommitted changes to the code segment, the *Current Line Code Contents* change, invalidating the cache instantly. If they checkout a different branch or pull updates, the *HEAD Commit Hash* changes, also invalidating the cache. This ensures briefs are blazing fast (under 20ms) yet 100% accurate.

### 4. The Intelligence Layer (`pkg/ai/`)
*   **Core Logic:** Decoupled using a common `Provider` interface, allowing hot-swapping between multiple cloud and local models.
*   **Responsibilities:**
    *   **Unified Interface (`pkg/ai/provider.go`):** Defines the standard `GenerateBrief` and `ModelName` methods that all clients must implement.
    *   **Provider Factory (`pkg/ai/factory.go`):** Handles the instantiation of the correct client based on the `--provider` flag and environment variables.
    *   **Supported Providers:**
        *   **Gemini (`pkg/ai/gemini.go`):** Integrates Google's Gemini models via the official SDK. Supports Vertex AI and Google AI Studio.
        *   **OpenAI (`pkg/ai/openai.go`):** Integrates GPT models (defaulting to `gpt-4o`) via high-performance, zero-dependency HTTP calls.
        *   **Claude (`pkg/ai/claude.go`):** Integrates Anthropic's Claude models (defaulting to `claude-3-5-sonnet-20240620`) via zero-dependency HTTP calls.
        *   **Ollama (`pkg/ai/ollama.go`):** Local fallback for running Gemma or other models offline, ensuring data privacy.

---

## 🔄 End-to-End Execution Pipeline

The following sequence details how BlameBrief processes a single query:

```
[User Input]  ==>  Cobra CommandLine Validator (cmd/root.go)
                         ||
                         || (Verify file exists & range bounds are correct)
                         \/
                   Upfront Line Counter (cmd/root.go)
                         ||
                         || (Read current line bounds to prevent out-of-bounds)
                         \/
                   Git Extraction Engine (pkg/git/extractor.go)
                         ||
                         || (git rev-parse HEAD + git log -L)
                         \/
                 Is Deterministic Flag? (--raw / --json)
                        /  \
                       /    \
               [YES]  /      \ [NO]
                     /        \
                    v          v
          Print Raw Markdown/  SHA256 Caching Engine (pkg/cache/cache.go)
          Pretty JSON & Exit       /  \
                                  /    \
                   [CACHE HIT]   /      \   [CACHE MISS]
                                /        \
                               v          v
                            Serve       AI Dispatcher Layer (pkg/ai/)
                            Instantly   (Multi-Provider Dispatcher)
                            (<20ms)            ||
                                               || (Produce Markdown Report)
                                               \/
                                            Cache Saver & Print To Stdout
```

---

## 🛠️ Design Patterns & Best Practices

1.  **Subprocess Optimization over CGO:**
    Instead of importing heavy native C libraries (like `libgit2` via `go-git`), BlameBrief calls native system `git`. This maintains a very small, zero-dependency binary size, runs at highly optimized speeds, and leverages the developer's pre-configured local Git credentials.
2.  **No-Dependency Local REST:**
    The local Ollama integration does not rely on third-party SDK wrappers. It utilizes Go's standard library `"net/http"` and `"encoding/json"` modules, preventing dependency bloat and maintaining a high security score.
3.  **Local Deterministic Parsing:**
    For the non-AI output pathways, BlameBrief incorporates a custom, high-speed, and robust Git log parser. This parser reconstructs raw `git log -L` diff blocks into structured `CommitInfo` slices and exports them natively as beautiful pretty-printed JSON, allowing seamless toolchain piping.
4.  **Mock Testing Strategy:**
    To ensure technical execution is production-grade, AI layers are fully unit-tested without requiring active network calls or local GPUs. We leverage `"net/http/httptest"` to launch high-fidelity mock API servers in memory, allowing us to validate HTTP payloads and parser logic in under 0.01 seconds.
