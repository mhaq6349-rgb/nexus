package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/muhammad/nexus/go/internal/runtime"
	"github.com/muhammad/nexus/go/internal/scheduler"
	"github.com/muhammad/nexus/go/internal/types"
)

type Server struct {
	sched   *scheduler.Scheduler
	rt      *runtime.RuntimeManager
	httpSrv *http.Server
}

func New(sched *scheduler.Scheduler, rt *runtime.RuntimeManager) *Server {
	return &Server{sched: sched, rt: rt}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/call", s.handleCall)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/runtimes", s.handleRuntimes)
	mux.HandleFunc("/api/v1/functions", s.handleFunctions)
	mux.HandleFunc("/", s.handleCORS(s.handleDocs))

	s.httpSrv = &http.Server{Addr: addr, Handler: mux}
	log.Printf("Nexus API listening on %s", addr)
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.httpSrv != nil {
		return s.httpSrv.Close()
	}
	return nil
}

func (s *Server) handleCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { w.WriteHeader(204); return }
		next(w, r)
	}
}

func (s *Server) json(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.json(w, 405, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Function string          `json:"function"`
		Args     []float64       `json:"args"`
		Raw      json.RawMessage `json:"raw"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.json(w, 400, map[string]string{"error": fmt.Sprintf("bad request: %v", err)})
		return
	}
	args := make([]types.Value, len(req.Args))
	for i, a := range req.Args {
		args[i] = types.ValF64From(a)
	}
	start := time.Now()
	result, err := s.sched.Call(req.Function, args, 10*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		s.json(w, 400, map[string]interface{}{
			"error": err.Error(), "function": req.Function, "elapsed_ms": elapsed.Milliseconds(),
		})
		return
	}
	s.json(w, 200, map[string]interface{}{
		"result": result, "function": req.Function, "elapsed_ms": elapsed.Milliseconds(),
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.json(w, 200, s.sched.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.json(w, 200, map[string]string{"status": "ok", "version": "0.1.0"})
}

func (s *Server) handleRuntimes(w http.ResponseWriter, r *http.Request) {
	s.json(w, 200, s.rt.CheckAvailability())
}

func (s *Server) handleFunctions(w http.ResponseWriter, r *http.Request) {
	s.json(w, 200, s.sched.Stats()["functions"])
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Nexus</title>
<style>body{font-family:system-ui;max-width:800px;margin:40px auto;padding:0 20px}
pre{background:#f5f5f5;padding:1em;border-radius:8px}
code{background:#eee;padding:2px 6px;border-radius:3px}
h1{color:#2563eb}</style></head><body>
<h1>⚡ Nexus</h1>
<p>Cross-language runtime bridge. Call functions across Go, Rust, Python, and TypeScript.</p>
<h2>Endpoints</h2>
<ul>
<li><code>POST /api/v1/call</code> — call a registered function</li>
<li><code>GET /api/v1/stats</code> — scheduler statistics</li>
<li><code>GET /api/v1/health</code> — health check</li>
<li><code>GET /api/v1/runtimes</code> — available runtimes</li>
<li><code>GET /api/v1/functions</code> — registered functions</li>
</ul>
<h2>Example</h2>
<pre>curl -X POST http://localhost:8080/api/v1/call \
  -H 'Content-Type: application/json' \
  -d '{"function":"math.add","args":[1,2,3]}'</pre>
<h2>Architecture</h2>
<pre>Go (scheduler) → Rust (SIMD transform) → Python (analytics) → TS (visualization)
       ↕ HTTP/JSON API (external)</pre>
</body></html>`)
}
