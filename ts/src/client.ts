/**
 * Nexus TypeScript client — communicates with the Go daemon via HTTP/WebSocket.
 */

export interface CallPayload {
  function: string;
  args: number[];
}

export interface CallResponse {
  result: number | string | Record<string, unknown>;
  function: string;
  elapsed_ms: number;
  error?: string;
}

export interface RuntimeStatus {
  go: boolean;
  rust: boolean;
  python: boolean;
  node: boolean;
}

export class NexusClient {
  private baseUrl: string;
  private wsUrl: string;
  private ws: WebSocket | null = null;
  private pending = new Map<number, { resolve: (v: unknown) => void; reject: (e: Error) => void }>();
  private callId = 0;

  constructor(config: { baseUrl?: string; wsUrl?: string } = {}) {
    this.baseUrl = config.baseUrl ?? 'http://localhost:8080';
    this.wsUrl = config.wsUrl ?? this.baseUrl.replace(/^http/, 'ws');
  }

  async call(functionName: string, ...args: number[]): Promise<CallResponse> {
    const payload: CallPayload = { function: functionName, args };
    const resp = await fetch(`${this.baseUrl}/api/v1/call`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({ error: resp.statusText }));
      throw new Error(`Nexus call error: ${err.error ?? resp.statusText}`);
    }
    return resp.json() as Promise<CallResponse>;
  }

  async stats(): Promise<Record<string, unknown>> {
    const resp = await fetch(`${this.baseUrl}/api/v1/stats`);
    return resp.json() as Promise<Record<string, unknown>>;
  }

  async health(): Promise<{ status: string; version: string }> {
    const resp = await fetch(`${this.baseUrl}/api/v1/health`);
    return resp.json() as Promise<{ status: string; version: string }>;
  }

  async runtimes(): Promise<RuntimeStatus> {
    const resp = await fetch(`${this.baseUrl}/api/v1/runtimes`);
    return resp.json() as Promise<RuntimeStatus>;
  }

  connectWebSocket(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(`${this.wsUrl}/ws`);
      this.ws.onopen = () => resolve();
      this.ws.onerror = (e) => reject(e);
      this.ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.call_id !== undefined) {
            const pending = this.pending.get(msg.call_id);
            if (pending) {
              if (msg.error) pending.reject(new Error(msg.error));
              else pending.resolve(msg.result);
              this.pending.delete(msg.call_id);
            }
          }
        } catch { /* ignore parse errors */ }
      };
    });
  }

  async callWs(functionName: string, ...args: number[]): Promise<unknown> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      await this.connectWebSocket();
    }
    return new Promise((resolve, reject) => {
      const callId = ++this.callId;
      this.pending.set(callId, { resolve, reject });
      this.ws!.send(JSON.stringify({ call_id: callId, function: functionName, args }));
      setTimeout(() => {
        this.pending.delete(callId);
        reject(new Error(`WebSocket call timeout: ${functionName}`));
      }, 10000);
    });
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
