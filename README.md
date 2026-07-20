# ⚡ Nexus — Cross-Language Runtime Bridge

**Go · Rust · Python · TypeScript — all in one process, zero serialization overhead.**

Nexus lets you write one pipeline across 4 languages and run it as a single system. Go schedules, Rust transforms (SIMD), Python analyzes (NumPy), TypeScript visualizes — all communicating through a lock-free shared memory ring buffer.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      HTTP/WS API (:8080)                       │
│  POST /call   GET /stats   GET /health   GET /runtimes        │
└──┬───────────────────────────────┬───────────────────────────┘
   │                               │
┌──▼──────────────┐    ┌──────────▼──────────┐
│  Go Daemon      │    │  TypeScript Client   │
│  ● Scheduler    │    │  ● Dashboard (React) │
│  ● HTTP API     │◄──►│  ● WASM Fallback    │
│  ● Lifecycle    │    │  ● Vite Dev Plugin   │
│  ● 8+ builtins  │    │  ● WebSocket Stream  │
└──┬──────────────┘    └─────────────────────┘
   │
┌──▼──────────────┐    ┌──────────────────────┐
│  Rust Core      │    │  Python Bridge       │
│  ● Shared Mem   │◄──►│  ● @nexus.export     │
│  ● Ring Buffer  │    │  ● NumPy Bridge      │
│  ● SIMD (wide)  │    │  ● ctypes FFI        │
│  ● C ABI Exports│    │  ● HTTP Client       │
└─────────────────┘    └──────────────────────┘
```

## Why Nexus?

| Problem | Nexus Solution |
|---------|---------------|
| Python slow at loops? | Rust handles the hot path via shared library C ABI |
| Go slow at SIMD? | Rust does vectorized math, Go orchestrates |
| TypeScript can't call Python? | HTTP/WS bridge with typed client |
| Cross-language data copying? | **Zero-copy** shared memory arena |

## Project Structure

```
nexus/
├── rust/          # Core shared library — memory, types, SIMD, FFI
│   ├── src/memory.rs    # Lock-free ring buffer + shared arena
│   ├── src/types.rs     # Cross-language type system
│   ├── src/ffi.rs       # C ABI exports for Go/Python
│   └── src/bridge.rs    # Function registry + dispatcher
├── go/            # Scheduler daemon — HTTP API, lifecycle
│   ├── cmd/nexusd/      # Daemon entrypoint
│   └── internal/        # Scheduler, runtime manager, types, API
├── python/        # Python client — export decorator, numpy bridge
│   └── src/nexus/       # Bridge, types, NumPy, export decorator
├── ts/            # TypeScript client — dashboard, WASM, Vite plugin
│   └── src/             # Client, WASM bridge, Vite plugin
├── examples/      # Cross-language pipeline demo
└── tests/         # Integration + performance tests
```

## Getting Started

### Prerequisites
- Go 1.22+ · Rust 1.75+ · Python 3.10+ · Node.js 22+

### Quick Start
```bash
# 1. Start the Nexus daemon
cd go && go run ./cmd/nexusd &

# 2. Run the Python analytics pipeline
cd python && python examples/pipeline.py

# 3. Run the TypeScript dashboard
cd ts && npx tsx examples/dashboard.ts

# 4. Build Rust for maximum performance
cd rust && cargo build --release
export NEXUS_RUST_LIB="$PWD/target/release/libnexus_core.so"
```

### HTTP API
```bash
# Call a function
curl -X POST http://localhost:8080/api/v1/call \
  -H 'Content-Type: application/json' \
  -d '{"function":"math.add","args":[1,2,3]}'

# Check health
curl http://localhost:8080/api/v1/health

# View statistics
curl http://localhost:8080/api/v1/stats
```

## Built-in Functions

| Function | Lang | Description |
|----------|------|-------------|
| `math.add` | Go | Sum of arguments |
| `math.mul` | Go | Product of arguments |
| `string.reverse` | Go | Reverse a string |
| `string.concat` | Go | Concatenate strings |
| `system.ping` | Go | Health check |
| `system.echo` | Go | Echo input |
| `data.generate` | Go | Generate test data |
| `http.fetch` | Go | Mock HTTP fetch |
| `analytics.summary` | Py | NumPy statistical summary |
| `analytics.fft` | Py | FFT frequency analysis |
| `analytics.transform` | Py | SIMD-style array transform |

## Agents & Forge

Built using the **Forge agent system** (152 agents across 118 skills):

| Agent | Role |
|-------|------|
| `architect-system` | System design & architecture |
| `backend` (api/db) | HTTP API & data layer |
| `rust` | SIMD, FFI, memory management |
| `go` | Scheduler, HTTP server, lifecycle |
| `python` | NumPy bridge, export decorators |
| `typescript` | Client SDK, WASM, Vite plugin |
| `devops-ci` | CI/CD pipeline (4 languages) |
| `security` | Input validation, type safety |

## License

MIT
