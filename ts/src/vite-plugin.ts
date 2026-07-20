/**
 * Vite plugin for Nexus — provides dev-time integration with the Nexus daemon.
 */

import type { Plugin, ViteDevServer } from 'vite';

export interface ViteNexusPlugin {
  (config?: NexusPluginConfig): Plugin;
}

export interface NexusPluginConfig {
  daemonUrl?: string;
  autoStart?: boolean;
  functions?: string[];
}

export function nexusVitePlugin(config: NexusPluginConfig = {}): Plugin {
  const daemonUrl = config.daemonUrl ?? 'http://localhost:8080';
  const functions = config.functions ?? [];

  return {
    name: 'nexus-vite-plugin',
    enforce: 'pre',

    async configureServer(server: ViteDevServer) {
      if (config.autoStart !== false) {
        server.config.logger.info(`⚡ Nexus daemon: ${daemonUrl}`);
        if (functions.length > 0) {
          server.config.logger.info(`   Registered functions: ${functions.join(', ')}`);
        }
      }

      server.ws.on('nexus:call', async (data: unknown) => {
        const msg = data as { function: string; args: number[]; call_id: number };
        try {
          const resp = await fetch(`${daemonUrl}/api/v1/call`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ function: msg.function, args: msg.args }),
          });
          const result = await resp.json();
          server.ws.send('nexus:result', { call_id: msg.call_id, ...result });
        } catch (err) {
          server.ws.send('nexus:result', {
            call_id: msg.call_id,
            error: String(err),
          });
        }
      });
    },

    transform(code: string, id: string) {
      if (!id.endsWith('.ts') && !id.endsWith('.tsx')) return;
      if (code.includes('nexus:function')) {
        return {
          code: code.replace(
            /nexus:function\s+(\w+)/g,
            (_, name) => `/* nexus:${name} */ ${name}`,
          ),
          map: null,
        };
      }
    },
  };
}
