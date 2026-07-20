package runtime

import (
	"fmt"
	"os/exec"
	"sync"
)

type LangRuntime int

const (
	RuntimeGo     LangRuntime = 0
	RuntimeRust   LangRuntime = 1
	RuntimePython LangRuntime = 2
	RuntimeTS     LangRuntime = 3
)

func (r LangRuntime) String() string {
	switch r {
	case RuntimeGo: return "go"
	case RuntimeRust: return "rust"
	case RuntimePython: return "python"
	case RuntimeTS: return "typescript"
	default: return "unknown"
	}
}

type RuntimeManager struct {
	mu       sync.RWMutex
	pythonPath string
	nodePath  string
}

func New(pythonPath, nodePath string) *RuntimeManager {
	if pythonPath == "" { pythonPath = "python" }
	if nodePath == "" { nodePath = "node" }
	return &RuntimeManager{pythonPath: pythonPath, nodePath: nodePath}
}

func (rm *RuntimeManager) ExecPython(script string, args ...string) (string, error) {
	cmd := exec.Command(rm.pythonPath, append([]string{"-c", script}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("python exec: %w\n%s", err, out)
	}
	return string(out), nil
}

func (rm *RuntimeManager) ExecNode(script string, args ...string) (string, error) {
	cmd := exec.Command(rm.nodePath, append([]string{"-e", script}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("node exec: %w\n%s", err, out)
	}
	return string(out), nil
}

func (rm *RuntimeManager) RunFile(lang LangRuntime, filepath string, args ...string) (string, error) {
	var cmd *exec.Cmd
	switch lang {
	case RuntimePython:
		cmd = exec.Command(rm.pythonPath, append([]string{filepath}, args...)...)
	case RuntimeTS:
		cmd = exec.Command(rm.nodePath, append([]string{filepath}, args...)...)
	case RuntimeRust:
		cmd = exec.Command("cargo", append([]string{"run", "--manifest-path", filepath}, args...)...)
	default:
		return "", fmt.Errorf("unsupported runtime: %s", lang)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("run %s: %w\n%s", lang, err, out)
	}
	return string(out), nil
}

func (rm *RuntimeManager) CheckAvailability() map[string]bool {
	return map[string]bool{
		"go":     true,
		"rust":   commandExists("cargo"),
		"python": commandExists(rm.pythonPath),
		"node":   commandExists(rm.nodePath),
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
