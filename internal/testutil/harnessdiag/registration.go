package harnessdiag

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"path/filepath"
	"sync"
)

// GraphRegistration shares the receipt owner's exact private binding checks.
// Updates touch only private coordination; they never write scanned evidence.
type GraphRegistration struct {
	mu           sync.Mutex
	root, file   string
	owner, value object
}

// Required registration poisons the invocation if any graph cannot join it.
// The closed sentinel intentionally prevents the wrapper roster from closing;
// a nested caller cannot catch this failure and leave a partial success claim.
func RegisterRequiredGraph(root, runRoot string, manifest map[string]any) (*GraphRegistration, error) {
	g := RegisterGraph(root, runRoot, manifest)
	if g != nil {
		return g, nil
	}
	if _, err := wrapperOwner(root); err == nil {
		_ = harnesslocal.PublishNew(filepath.Join(root, "registration-failed.json"), []byte(`{"incomplete":true}`))
	}
	return nil, errArtifact
}

func RegisterGraph(root, runRoot string, manifest map[string]any) *GraphRegistration {
	owner, err := wrapperOwner(root)
	if err != nil {
		return nil
	}
	graphs, err := wrapperGraphs(root, owner)
	if err != nil || len(graphs) >= 64 {
		return nil
	}
	value := object{"version": 1, "invocation_digest": digest([]byte(owner["invocation_id"].(string))), "run_root": runRoot, "run_id": manifest["run_id"], "source_digest": manifest["source_digest"], "graph_digest": manifest["graph_digest"], "phase": "registered", "scan_start_ms": nil, "scan_end_ms": nil, "scan_outcome": nil}
	file := filepath.Join(root, "graph-"+digest([]byte(runRoot))[7:]+".json")
	bytes, err := json.Marshal(value)
	if err != nil || harnesslocal.PublishNew(file, bytes) != nil {
		return nil
	}
	if _, err := wrapperGraphs(root, owner); err != nil {
		return nil
	}
	return &GraphRegistration{root: root, file: file, owner: owner, value: value}
}
func (g *GraphRegistration) Digest() any {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value["invocation_digest"]
}
func (g *GraphRegistration) update(operation string, passed bool) {
	if g == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	owner, err := wrapperOwner(g.root)
	if err != nil || owner["invocation_id"] != g.owner["invocation_id"] || g.value["phase"] == "complete" {
		return
	}
	// Check this process's time namespace against the wrapper before recording
	// an instant in its clock. A mismatched namespace leaves time unavailable.
	var at any
	if clockCompatible(owner) {
		if n, err := bootMilliseconds(); err == nil {
			at = n
		}
	}
	switch operation {
	case "start":
		if g.value["phase"] != "registered" {
			return
		}
		g.value["phase"] = "scanning"
		g.value["scan_start_ms"] = at
	case "finish":
		if g.value["phase"] != "scanning" {
			return
		}
		g.value["phase"] = "scanned"
		g.value["scan_end_ms"] = at
		g.value["scan_outcome"] = "failed"
		if passed {
			g.value["scan_outcome"] = "passed"
		}
	case "complete":
		g.value["phase"] = "complete"
	default:
		return
	}
	bytes, err := json.Marshal(g.value)
	if err == nil {
		_ = harnesslocal.ReplaceFile(g.file, bytes)
	}
}
func (g *GraphRegistration) ScanStarted()             { g.update("start", false) }
func (g *GraphRegistration) ScanFinished(passed bool) { g.update("finish", passed) }
func (g *GraphRegistration) Complete()                { g.update("complete", false) }
