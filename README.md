# Clanker Rehab

🦝 **Thinking is BACK**

An interactive learning platform to master Go fundamentals without AI assistance. Features 104 structured concepts, spaced repetition, and a brushtail possum.

## Features

- **104 Go Concepts** across 3 difficulty levels (Beginner, Intermediate, Advanced)
- **Difficulty Filters** - Toggle between Beginner, Intermediate, and Advanced concepts
- **Ordered Categories** - Core Syntax first, followed by importance-based ordering
- **CodeMirror Editor** with Go syntax highlighting and Monokai theme
- **Spaced Repetition** - concepts expire and return to practice queue
- **Local Code Execution** - runs your Go code safely in isolated temp directories
- **No Dependencies** - fully self-contained, runs on localhost
- **Progress Tracking** - localStorage tracks learned concepts and expiry timers
- **Improved Instructions** - Clear, detailed guidance for each concept

## Quick Start

```bash
cd go-concept-trainer
go run main.go
```

Open your browser to: **http://localhost:3000**

## How to Use

1. **Difficulty Filters** (Top of left panel):
   - **Green Button (Beginner)**: Basic Go syntax and fundamentals
   - **Orange Button (Intermediate)**: More complex patterns and techniques
   - **Burgundy Button (Advanced)**: Generics, reflection, advanced concurrency
   - Click to toggle filters (multiple can be active)

2. **Left Panel**: Click any concept card to load it into the editor
   - Categories ordered by importance (Core Syntax → Data Structures → etc.)
   - Shows concept count per category

3. **Center Panel**:
   - Read the detailed instruction
   - Write Go code to satisfy the requirement
   - Click "Run Code" to validate
   - Click "Reset" to restore boilerplate

4. **Right Panel**: Successfully completed concepts appear here with expiry countdown

5. **Settings**: Configure default expiry time (default: 14 days)

## Concept Categories (104 Total)

**Difficulty Breakdown:**
- 🟢 Beginner: 87 concepts
- 🟠 Intermediate: 5 concepts
- 🔴 Advanced: 12 concepts

**By Category:**
- **Core Syntax** (13 concepts): variables, constants, loops, conditionals, iota with bitmasks
- **Data Structures** (16 concepts): arrays, slices, maps, structs, capacity vs length
- **Functions & Closures** (11 concepts): parameters, returns, variadic, defer, method expressions
- **Pointers & Methods** (9 concepts): pointers, receivers, mutation, method overrides
- **Interfaces** (9 concepts): definition, type assertions, Stringer, type constraints for generics
- **Concurrency** (14 concepts): goroutines, channels, select, WaitGroup, Mutex, context, atomics
- **Standard Library** (14 concepts): fmt, strings, time, json, errors, sort, bufio
- **Error Handling** (6 concepts): custom errors, wrapping, panic/recover, errors.Is/As
- **Tooling & Tests** (3 concepts): packages, imports, aliases
- **Miscellaneous** (9 concepts): init, embedding, zero values, generics, reflection

## How It Works

### Validation
- Your code runs via `go run` in an isolated temp directory
- Output (stdout) is compared against expected output
- Success → concept moves to "Learned" panel with timer
- Failure → stays in practice queue

### Expiry & Spaced Repetition
- Each learned concept has a configurable expiry (default: 14 days)
- Timers update every 60 seconds
- Expired concepts automatically return to "Unlearned" queue
- Customize expiry duration in Settings

### Safety
- 5-second timeout prevents infinite loops
- Temp directory isolation
- Local-only execution (no network access)
- No persistent storage or authentication

## Project Structure

```
go-concept-trainer/
├── main.go              # HTTP server
├── concepts/            # Individual concept files (104 total)
│   ├── types.go         # Concept type definitions
│   ├── 001_var_declaration.go
│   ├── 002_short_declaration.go
│   ├── ...
│   └── 104_slice_capacity.go
├── templates/
│   └── index.html       # Single-page UI
├── static/
│   ├── script.js        # Frontend logic
│   ├── style.css        # Three-column layout
│   └── codemirror/      # CodeMirror editor assets
└── go.mod
```

### Concept Files

Each concept is now in its own numbered file (LeetCode-style numbering):
- **Format**: `XXX_concept-name.go` (e.g., `001_var_declaration.go`)
- **Numbering**: 001-104, ordered by category and difficulty
- **Structure**: Each file contains a single `ConceptXXX` variable and registers it via `init()`

This makes it easy to:
- Find specific concepts quickly
- Add new concepts without modifying a huge array
- See concept numbers in the UI (like LeetCode problems)
- Manage and test concepts individually

## Requirements

- Go 1.21+
- Modern web browser with localStorage support

## License

MIT

<!-- Auto-deploy test Thu Nov 13 02:53:54 AEDT 2025 -->
<!-- Test 1762962972 -->
<!-- Test 1762963005 -->
<!-- Final test 1762963086 -->
