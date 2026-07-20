package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/muhammad/nexus/internal/types"
)

type Scheduler struct {
	mu         sync.RWMutex
	functions  map[string]*types.FunctionInfo
	callID     atomic.Uint64
	pending    map[uint64]chan types.Value
	startedAt  time.Time
	totalCalls atomic.Int64
}

func New() *Scheduler {
	return &Scheduler{
		functions: make(map[string]*types.FunctionInfo),
		pending:   make(map[uint64]chan types.Value),
		startedAt: time.Now(),
	}
}

func (s *Scheduler) Register(name string, lang uint8, handler types.FunctionHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.functions[name] = &types.FunctionInfo{
		Name: name, Handler: handler, Lang: lang,
	}
}

func (s *Scheduler) Call(name string, args []types.Value, timeout time.Duration) (types.Value, error) {
	s.mu.RLock()
	info, ok := s.functions[name]
	s.mu.RUnlock()
	if !ok {
		return types.ValNullValue(), fmt.Errorf("function %q not registered", name)
	}
	callID := s.callID.Add(1)
	ch := make(chan types.Value, 1)
	s.mu.Lock()
	s.pending[callID] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.pending, callID)
		s.mu.Unlock()
	}()

	s.totalCalls.Add(1)
	go func() {
		start := time.Now()
		result, err := info.Handler(args)
		info.Latency = time.Since(start)
		info.Count++
		if err != nil {
			ch <- types.ValStr(fmt.Sprintf("error: %v", err))
		} else {
			ch <- result
		}
	}()

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(timeout):
		return types.ValNullValue(), fmt.Errorf("call %q timed out after %v", name, timeout)
	}
}

func (s *Scheduler) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := make(map[string]interface{})
	funcs := make([]map[string]interface{}, 0)
	for _, info := range s.functions {
		if info.Count > 0 {
			funcs = append(funcs, map[string]interface{}{
				"name": info.Name, "lang": info.Lang,
				"count": info.Count, "avg_latency_us": info.Latency.Microseconds(),
			})
		}
	}
	stats["functions"] = funcs
	stats["total_calls"] = s.totalCalls.Load()
	stats["uptime_seconds"] = int(time.Since(s.startedAt).Seconds())
	return stats
}

func (s *Scheduler) InitDefault() {
	s.Register("math.add", types.LangGo, func(args []types.Value) (types.Value, error) {
		var sum float64
		for _, a := range args {
			if f, ok := a.AsF64(); ok { sum += f }
		}
		return types.ValF64From(sum), nil
	})
	s.Register("math.mul", types.LangGo, func(args []types.Value) (types.Value, error) {
		prod := 1.0
		for _, a := range args {
			if f, ok := a.AsF64(); ok { prod *= f }
		}
		return types.ValF64From(prod), nil
	})
	s.Register("string.reverse", types.LangGo, func(args []types.Value) (types.Value, error) {
		s := ""
		if len(args) > 0 { s, _ = args[0].AsStr() }
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return types.ValStr(string(runes)), nil
	})
	s.Register("string.concat", types.LangGo, func(args []types.Value) (types.Value, error) {
		var result string
		for _, a := range args {
			if s, ok := a.AsStr(); ok { result += s }
		}
		return types.ValStr(result), nil
	})
	s.Register("system.echo", types.LangGo, func(args []types.Value) (types.Value, error) {
		if len(args) == 0 { return types.ValNullValue(), nil }
		return args[0], nil
	})
	s.Register("system.ping", types.LangGo, func(args []types.Value) (types.Value, error) {
		return types.ValStr("pong"), nil
	})
	s.Register("data.generate", types.LangGo, func(args []types.Value) (types.Value, error) {
		n := 1000
		if len(args) > 0 {
			if v, ok := args[0].AsI64(); ok { n = int(v) }
		}
		data := make([]byte, n*4)
		for i := 0; i < n; i++ {
			f := float64(i) * 1.5
			binary := make([]byte, 4)
			// manual float32 bits
			bits := uint32(f)
			data[i*4] = byte(bits)
			data[i*4+1] = byte(bits >> 8)
			data[i*4+2] = byte(bits >> 16)
			data[i*4+3] = byte(bits >> 24)
			_ = binary
		}
		return types.NewBytesVal(data), nil
	})
	s.Register("data.filter", types.LangGo, func(args []types.Value) (types.Value, error) {
		if len(args) < 1 { return types.ValNullValue(), fmt.Errorf("need data") }
		data, _ := args[0].AsStr()
		if len(args) > 1 { _, _ = args[1].AsF64() }

		return types.NewBytesVal([]byte(data)), nil
	})
	s.Register("http.fetch", types.LangGo, func(args []types.Value) (types.Value, error) {
		url := ""
		if len(args) > 0 { url, _ = args[0].AsStr() }
		return types.ValStr(fmt.Sprintf(`{"url":%q,"status":200,"body":"mock response for %s"}`, url, url)), nil
	})
}
