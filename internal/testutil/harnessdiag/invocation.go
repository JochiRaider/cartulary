package harnessdiag

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// OpenInvocation lets the native outer controller use the existing graph and
// receipt registration contract. It creates no independent graph registry.
func OpenInvocation(root string) error {
	directory, err := openDirectory(root, true)
	if err != nil {
		return err
	}
	directory.Close()
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return err
	}
	id, err := os.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		return err
	}
	stat, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return err
	}
	end := strings.LastIndex(string(stat), ") ")
	if end < 0 {
		return errArtifact
	}
	fields := strings.Fields(string(stat)[end+2:])
	if len(fields) < 20 {
		return errArtifact
	}
	ns, err := os.Readlink("/proc/self/ns/time")
	if err != nil {
		return err
	}
	start, err := bootMilliseconds()
	if err != nil {
		return err
	}
	owner := object{"version": 1, "invocation_id": strings.TrimSpace(string(id)), "pid": os.Getpid(), "start_ticks": fields[19], "start_ms": start, "boot_id": strings.TrimSpace(string(boot)), "namespace": ns}
	data, err := json.Marshal(owner)
	if err != nil {
		return err
	}
	if err = publishNew(filepath.Join(root, "owner.json"), data); err != nil {
		return err
	}
	_, err = wrapperOwner(root)
	return err
}

func InvocationIdentity(root string) (string, error) {
	owner, err := wrapperOwner(root)
	if err != nil {
		return "", err
	}
	return digest([]byte(owner["invocation_id"].(string))), nil
}

func CompleteInvocation(root string, exit, signal int) error {
	owner, err := wrapperOwner(root)
	if err != nil {
		return err
	}
	if owner["pid"] != float64(os.Getpid()) || !clockCompatible(owner) || exit < 0 || exit > 255 || (signal != 0 && signal != 2 && signal != 15) {
		return errArtifact
	}
	end, err := bootMilliseconds()
	if err != nil {
		return err
	}
	value := object{"end_ms": end, "exit_code": exit, "signal": signal, "boot_id": owner["boot_id"], "namespace": owner["namespace"]}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return publishNew(filepath.Join(root, "completion.json"), data)
}

// PublishInvocationAccounting adds only closed numeric facts to already-closed
// evidence. Every membership is checked before any graph receives publication.
func PublishInvocationAccounting(root string, metrics map[string]any) error {
	return publishInvocationAccounting(root, metrics, publishNew)
}

func publishInvocationAccounting(root string, metrics map[string]any, publish func(string, []byte) error) error {
	owner, err := wrapperOwner(root)
	if err != nil {
		return err
	}
	graphs, err := wrapperGraphs(root, owner)
	if err != nil || len(graphs) == 0 {
		return errArtifact
	}
	id := digest([]byte(owner["invocation_id"].(string)))
	source := graphs[0]["source_digest"]
	expected := map[string]any{}
	for _, graph := range graphs {
		expected[digest([]byte(graph["run_id"].(string)))] = graph["graph_digest"]
	}
	members := []any{}
	for _, graph := range graphs {
		if graph["phase"] != "complete" || graph["scan_outcome"] != "passed" {
			return errArtifact
		}
		cleanup, _, err := boundObject(filepath.Join(graph["run_root"].(string), "cleanup-results.json"), "cartulary.harness_cleanup_results.v2", 16*1024*1024)
		if err != nil || cleanup["run_id"] != graph["run_id"] {
			return errArtifact
		}
		for _, raw := range cleanup["results"].([]any) {
			if raw.(object)["outcome"] != "completed" {
				return errArtifact
			}
		}
		receipt, hash, err := boundObject(filepath.Join(graph["run_root"].(string), "diagnostics/terminal-receipt.json"), "cartulary.harness_terminal_receipt.v1", 65536)
		if err != nil {
			return err
		}
		if graph["source_digest"] != source || receipt["source_digest"] != source || receipt["invocation_digest"] != id || receipt["child_exit_code"] != metrics["child_exit_code"] {
			return errArtifact
		}
		roster := receipt["graph_memberships"].([]any)
		seen := map[string]bool{}
		for _, raw := range roster {
			member := raw.(object)
			run := member["run_digest"].(string)
			if seen[run] || expected[run] != member["graph_digest"] {
				return errArtifact
			}
			seen[run] = true
		}
		if len(seen) != len(expected) {
			return errArtifact
		}
		members = append(members, object{"run_digest": digest([]byte(graph["run_id"].(string))), "graph_digest": graph["graph_digest"], "receipt_digest": hash})
	}
	value := object{"schema_id": "cartulary.harness_invocation_accounting.v1", "invocation_digest": id, "source_digest": source, "graph_memberships": members, "metrics": metrics}
	data, err := json.Marshal(value)
	if err != nil || len(data) > 65535 {
		return errArtifact
	}
	normalized, err := Decode(data)
	if err != nil {
		return err
	}
	if err = Validate("cartulary.harness_invocation_accounting.v1", normalized); err != nil {
		return fmt.Errorf("accounting publication: %w", err)
	}
	for _, graph := range graphs {
		if err = publish(filepath.Join(graph["run_root"].(string), "diagnostics/invocation-accounting.json"), append(data, '\n')); err != nil {
			return err
		}
	}
	return nil
}
