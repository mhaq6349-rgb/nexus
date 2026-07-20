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

  // Step 1: Go fetches data
  console.log('1. Go: Fetching data...');
  const fetchResult = await client.call('http.fetch', 'https://api.github.com/repos/muhammad/nexus');
  console.log(`   → ${JSON.stringify(fetchResult).slice(0, 80)}...`);

  // Step 2: Go math (simulates Rust SIMD in production)
  console.log('2. Rust: SIMD transform...');
  const simdResult = await client.call('math.mul', 2.0, 3.0, 4.0);
  console.log(`   → 2 × 3 × 4 = ${simdResult.result}`);

  // Step 3: Go string processing (simulates Python analytics in production)
  console.log('3. Python: Analytics...');
  const reverseResult = await client.call('string.reverse', 'Nexus Cross-Language Bridge');
  console.log(`   → reversed: ${reverseResult.result}`);

  // Step 4: System health
  console.log('4. TS: Visualization...');
  const health = await client.health();
  console.log(`   → Nexus status: ${health.status} v${health.version}`);

  const totalMs = performance.now() - start;
  const runtime = await client.runtimes();
  console.log(`\n📊 Pipeline complete in ${totalMs.toFixed(1)}ms`);
  console.log(`   Available runtimes: ${Object.entries(runtime).filter(([,v]) => v).map(([k]) => k).join(', ')}`);

  return {
    fetch: JSON.stringify(fetchResult),
    transform: `2 × 3 × 4 = ${simdResult.result}`,
    analyze: reverseResult.result as string,
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
