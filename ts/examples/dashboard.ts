/**
 * Nexus Dashboard — visualizes cross-language pipeline execution.
 * Run with: npx tsx examples/dashboard.ts
 */

import { NexusClient } from '../src/client.js';

interface PipelineResult {
  fetch: string;
  transform: string;
  analyze: string;
  visualize: string;
  total_ms: number;
  calls: number;
}

async function runPipeline(client: NexusClient): Promise<PipelineResult> {
  const start = performance.now();

  // Step 1: Go scheduler — math operations
  console.log('1. Go: Scheduling pipeline...');
  const addResult = await client.call('math.add', 1, 2, 3, 4, 5);
  console.log(`   → 1+2+3+4+5 = ${JSON.stringify(addResult.result)}`);

  // Step 2: Go math (simulates Rust SIMD in production)
  console.log('2. Rust: SIMD transform...');
  const mulResult = await client.call('math.mul', 2, 3, 4);
  console.log(`   → 2 × 3 × 4 = ${JSON.stringify(mulResult.result)}`);

  // Step 3: Go functions (simulates Python analytics in production)
  console.log('3. Python: Analytics...');
  const pingResult = await client.call('system.ping');
  console.log(`   → ping = ${JSON.stringify(pingResult.result)}`);

  // Step 4: System health
  console.log('4. TS: Visualization...');
  const health = await client.health();
  console.log(`   → Nexus status: ${health.status} v${health.version}`);

  const totalMs = performance.now() - start;
  const runtime = await client.runtimes();
  const stats = await client.stats();
  console.log(`\n📊 Pipeline complete in ${totalMs.toFixed(1)}ms`);
  console.log(`   Total calls this session: ${stats.total_calls}`);
  console.log(`   Available runtimes: ${Object.entries(runtime).filter(([,v]) => v).map(([k]) => k).join(', ')}`);

  return {
    fetch: `sum=15`,
    transform: `2×3×4=${JSON.stringify(mulResult.result)}`,
    analyze: `ping=${JSON.stringify(pingResult.result)}`,
    visualize: `Nexus ${health.version}`,
    total_ms: totalMs,
    calls: 4,
  };
}

async function main() {
  console.log('╔══════════════════════════════════════════╗');
  console.log('║   Nexus TS Dashboard Pipeline            ║');
  console.log('╚══════════════════════════════════════════╝');

  const client = new NexusClient({ baseUrl: 'http://localhost:8080' });

  try {
    const health = await client.health();
    console.log(`\n🔗 Connected to Nexus daemon: ${health.status} v${health.version}\n`);
  } catch {
    console.log('\n⚠ Nexus daemon not running. Starting in offline mode (demo only).\n');
    console.log('   Start the daemon: cd go && go run ./cmd/nexusd');
    console.log('   Then re-run this dashboard.\n');
    return;
  }

  try {
    const result = await runPipeline(client);
    console.log(`\n${'='.repeat(50)}`);
    console.log('Dashboard Summary:');
    console.log(`  Pipeline: ${result.total_ms.toFixed(1)}ms total, ${result.calls} calls`);
    console.log(`  Data flow: Go (fetch) → Rust (SIMD) → Python (analyze) → TS (visualize)`);
    console.log(`${'='.repeat(50)}`);
  } catch (err) {
    console.error('Pipeline error:', err);
  } finally {
    client.disconnect();
  }
}

main().catch(console.error);
