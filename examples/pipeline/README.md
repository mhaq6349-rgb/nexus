# Nexus Cross-Language Pipeline Demo

This demo runs a 4-stage pipeline across Go, Rust, Python, and TypeScript.

## Data Flow

```
Go (fetch data) ──→ Rust (SIMD transform) ──→ Python (numpy analysis) ──→ TS (dashboard)
       │                    │                        │                         │
       └────────────────────┴────────────────────────┴─────────────────────────┘
                                  Nexus Shared Memory Ring Buffer
```

## Stages

| Stage | Language | What it does |
|-------|----------|-------------|
| 1 | Go | Generates/fetches data, schedules pipeline, serves HTTP API |
| 2 | Rust | SIMD-accelerated transforms via shared library C ABI |
| 3 | Python | NumPy analytics (stats, FFT, transforms) via `@nexus.export` |
| 4 | TypeScript | Real-time dashboard, WASM fallback, Vite plugin |

## Run It

```bash
# 1. Start the Nexus daemon (Go scheduler)
cd go && go run ./cmd/nexusd &

# 2. Run Python analytics
cd python && python examples/pipeline.py

# 3. Run TS dashboard
cd ts && npx tsx examples/dashboard.ts

# 4. Build Rust shared library (optional, for SIMD)
cd rust && cargo build --release
# Set NEXUS_RUST_LIB=../rust/target/release/libnexus_core.so
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     HTTP API (port 8080)                  │
│  POST /api/v1/call  GET /api/v1/stats  GET /api/v1/health│
└────────┬──────────────────────────────┬──────────────────┘
         │                              │
    ┌────▼────┐                   ┌─────▼─────┐
    │  Go daemon │                 │  TS Client  │
    │ schedule  │◄─── HTTP/WS ───►│  dashboard  │
    │ + funcs   │                 │  + WASM    │
    └────┬────┘                   └───────────┘
         │
    ┌────▼────┐
    │ Rust FFI │◄── ctypes/cgo ──► Python client
    │ SIMD lib │                  │ @export    │
    └─────────┘                  └────────────┘
```
