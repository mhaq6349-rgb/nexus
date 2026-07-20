package main

import (
	"fmt"
	"time"

	"github.com/muhammad/nexus/go/internal/scheduler"
	"github.com/muhammad/nexus/go/internal/types"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   Nexus Cross-Language Pipeline Demo     ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	sched := scheduler.New()
	sched.InitDefault()

	// Step 1: Generate data
	fmt.Println("\n1. Go: Generating data...")
	data, err := sched.Call("data.generate", []types.Value{types.ValI64From(100)}, 5*time.Second)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
		return
	}
	fmt.Printf("   Generated data: %v bytes\n", len(data.Bytes))

	// Step 2: Math operations
	fmt.Println("\n2. Go: Math operations...")
	sum, _ := sched.Call("math.add", []types.Value{types.ValF64From(10), types.ValF64From(20), types.ValF64From(30)}, 5*time.Second)
	prod, _ := sched.Call("math.mul", []types.Value{types.ValF64From(2), types.ValF64From(3), types.ValF64From(4)}, 5*time.Second)
	fmt.Printf("   10 + 20 + 30 = %.0f\n", sum.F64)
	fmt.Printf("   2 × 3 × 4 = %.0f\n", prod.F64)

	// Step 3: String processing
	fmt.Println("\n3. Go: String processing...")
	reversed, _ := sched.Call("string.reverse", []types.Value{types.ValStr("Nexus Bridge")}, 5*time.Second)
	concat, _ := sched.Call("string.concat", []types.Value{types.ValStr("Hello"), types.ValStr(" "), types.ValStr("World")}, 5*time.Second)
	fmt.Printf("   'Nexus Bridge' reversed = %s\n", reversed.String)
	fmt.Printf("   Concatenated = %s\n", concat.String)

	// Step 4: System
	fmt.Println("\n4. Go: System functions...")
	ping, _ := sched.Call("system.ping", nil, 5*time.Second)
	httpResp, _ := sched.Call("http.fetch", []types.Value{types.ValStr("https://api.github.com")}, 5*time.Second)
	fmt.Printf("   Ping = %s\n", ping.String)
	fmt.Printf("   HTTP fetch = %s\n", httpResp.String)

	// Stats
	fmt.Println("\n5. Pipeline complete.")
	fmt.Println("\n📊 Statistics:")
	stats := sched.Stats()
	fmt.Printf("   Total calls: %d\n", stats["total_calls"])
	fmt.Printf("   Functions registered: %d\n", len(sched.Stats()["functions"].([]map[string]interface{})))
	fmt.Println("\n✓ Pipeline demonstrates Go as the scheduler/orchestrator.")
	fmt.Println("  In production, Rust handles SIMD transforms,")
	fmt.Println("  Python handles analytics, and TypeScript handles visualization.")
}
