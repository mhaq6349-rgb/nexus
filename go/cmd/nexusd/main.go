package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhammad/nexus/internal/api"
	"github.com/muhammad/nexus/internal/runtime"
	"github.com/muhammad/nexus/internal/scheduler"
)

var (
	port       = flag.String("port", ":8080", "HTTP server port")
	pythonPath = flag.String("python", "python", "Python executable path")
	nodePath   = flag.String("node", "node", "Node.js executable path")
)

func main() {
	flag.Parse()
	log.Println("Starting Nexus daemon...")

	sched := scheduler.New()
	sched.InitDefault()

	rt := runtime.New(*pythonPath, *nodePath)

	avail := rt.CheckAvailability()
	log.Printf("Runtime availability: Go=✓ Rust=%v Python=%v Node=%v",
		avail["rust"], avail["python"], avail["node"])

	srv := api.New(sched, rt)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		srv.Stop()
		os.Exit(0)
	}()

	if err := srv.Start(*port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
