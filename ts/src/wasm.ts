/**
 * Nexus WASM bridge — runs Rust-compiled modules in the browser.
 */

export interface WasmModule {
  memory: WebAssembly.Memory;
  nexus_init: () => number;
  nexus_apply_simd_f32: (input: number, output: number, len: number, op: number) => number;
}

export class NexusWasmBridge {
  private module: WasmModule | null = null;
  private heap: Float32Array | null = null;
  private heapPtr = 0;
  private heapSize = 0;

  async load(wasmUrl: string): Promise<void> {
    const response = await fetch(wasmUrl);
    const bytes = await response.arrayBuffer();
    const importObj = {
      env: {
        memory: new WebAssembly.Memory({ initial: 256, maximum: 512 }),
        emscripten_memcpy: (dest: number, src: number, num: number) => {
          if (this.heap) {
            this.heap.set(new Float32Array(this.heap.buffer, src, num / 4), dest / 4);
          }
        },
      },
    };
    const result = await WebAssembly.instantiate(bytes, importObj);
    this.module = result.instance.exports as unknown as WasmModule;
    this.heap = new Float32Array(this.module.memory.buffer);
    this.heapSize = this.heap.length;
    this.module.nexus_init();
  }

  get loaded(): boolean {
    return this.module !== null;
  }

  simdMul(data: Float32Array, scalar: number): Float32Array {
    if (!this.module || !this.heap) {
      return data.map(v => v * scalar);
    }
    const n = data.length;
    if (this.heapPtr + n * 4 > this.heapSize) {
      this.heapPtr = 0;
    }
    this.heap.set(data, this.heapPtr / 4);
    const inPtr = this.heapPtr;
    const outPtr = this.heapPtr + n * 4;
    this.module.nexus_apply_simd_f32(inPtr, outPtr, n, 1);
    const result = new Float32Array(this.heap.buffer, outPtr, n);
    this.heapPtr = outPtr + n * 4;
    return result;
  }

  relu(data: Float32Array): Float32Array {
    if (!this.module || !this.heap) {
      return data.map(v => Math.max(0, v));
    }
    const n = data.length;
    if (this.heapPtr + n * 4 > this.heapSize) {
      this.heapPtr = 0;
    }
    this.heap.set(data, this.heapPtr / 4);
    const inPtr = this.heapPtr;
    const outPtr = this.heapPtr + n * 4;
    this.module.nexus_apply_simd_f32(inPtr, outPtr, n, 2);
    const result = new Float32Array(this.heap.buffer, outPtr, n);
    this.heapPtr = outPtr + n * 4;
    return result;
  }

  dispose(): void {
    this.module = null;
    this.heap = null;
  }
}
