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
| (Extraction) |               |   (Caching)    |        | (Gemini/Oll.) |
+--------------+               +----------------+        +---------------+
```

---

## 🗂️ Tech Layering & Packages

The codebase is organized into four main layers:

### 1. The Orchestration Layer (`cmd/`)
*   **Core Logic:** Defined in `cmd/root.go` using `spf13/cobra`.
*   **Responsibilities:**
    *   Exposes and parses CLI arguments and flags (`--lines`, `--detail`, `--local`, `--model`, `--clear-cache`).
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
*   **Core Logic:** Decoupled into `gemini.go` (Cloud) and `ollama.go` (Local Fallback).
*   **Responsibilities:**
    *   **Cloud Driver (`pkg/ai/gemini.go`):** Integrates the official, modern `google.golang.org/genai` Go SDK. It supports both **Google AI Studio** (using standard API keys) and **Google Cloud Vertex AI** (using GCP projects and region variables) via a unified interface. Defaults to **`gemini-3.5-flash`** for optimal speed and context length.
    *   **Local Driver (`pkg/ai/ollama.go`):** Integrates a lightweight REST client calling a local Ollama API server. This fallback supports running state-of-the-art **Gemma 4** models locally and offline, ensuring sensitive enterprise files never leave the machine.

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
                            Instantly   (Query Gemini 3.5 / Local Gemma 4)
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
