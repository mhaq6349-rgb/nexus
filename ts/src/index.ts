export { NexusClient } from './client.js';
export { NexusWasmBridge } from './wasm.js';
export { nexusVitePlugin } from './vite-plugin.js';

export interface NexusConfig {
  baseUrl?: string;
  wsUrl?: string;
  wasmUrl?: string;
}

export interface CallResult {
  result: unknown;
  function: string;
  elapsed_ms: number;
  error?: string;
}

export interface NexusStats {
  functions: Array<{
    name: string;
    lang: number;
    count: number;
    avg_latency_us: number;
  }>;
  total_calls: number;
  uptime_seconds: number;
}

export type { ViteNexusPlugin } from './vite-plugin.js';
