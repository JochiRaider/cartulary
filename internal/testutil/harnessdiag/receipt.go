package harnessdiag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func exactKeys(value object, keys ...string) bool {
	if len(value) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := value[key]; !ok {
			return false
		}
	}
	return true
}
func safeInteger(value any, min, max float64) bool {
	number, ok := value.(float64)
	return ok && number >= min && number <= max && math.Trunc(number) == number
}
func match(pattern string, value any) bool {
	text, ok := value.(string)
	return ok && schemaPattern(pattern, text)
}
func boundObject(file, schema string, maximum int64) (object, string, error) {
	data, err := readLocal(file, maximum, false)
	if err != nil {
		return nil, "", err
	}
	value, err := Decode(data)
	if err != nil {
		return nil, "", err
	}
	v, ok := value.(object)
	if !ok {
		return nil, "", errArtifact
	}
	if schema != "" {
		if err := Validate(schema, v); err != nil {
			return nil, "", err
		}
	}
	return v, digest(data), nil
}
func bootMilliseconds() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	return parseBootMilliseconds(string(data))
}

func parseBootMilliseconds(value string) (float64, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0, errArtifact
	}
	// /proc/uptime has hundredth-second precision. Parsing through binary
	// floating point can yield fractional milliseconds and invalidate receipts.
	parts := strings.Split(fields[0], ".")
	if len(parts) != 2 || len(parts[1]) != 2 || !regexp.MustCompile(`^[0-9]+\.[0-9]{2}$`).MatchString(fields[0]) {
		return 0, errArtifact
	}
	seconds, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || seconds > 9007199254740 {
		return 0, errArtifact
	}
	hundredths, err := strconv.ParseUint(parts[1], 10, 64)
	milliseconds := seconds*1000 + hundredths*10
	if err != nil || milliseconds > 9007199254740991 {
		return 0, errArtifact
	}
	return float64(milliseconds), nil
}
func wrapperOwner(root string) (object, error) {
	directory, err := openDirectory(root, true)
	if err != nil {
		return nil, err
	}
	directory.Close()
	owner, _, err := boundObject(filepath.Join(root, "owner.json"), "", 65536)
	if err != nil {
		return nil, err
	}
	if !exactKeys(owner, "version", "invocation_id", "pid", "start_ticks", "start_ms", "boot_id", "namespace") || owner["version"] != float64(1) || !match(`^[a-f0-9-]{36}$`, owner["invocation_id"]) || !safeInteger(owner["pid"], 1, 9007199254740991) || !match(`^[0-9]+$`, owner["start_ticks"]) || !safeInteger(owner["start_ms"], 0, 9007199254740991) || !match(`^[a-f0-9-]{36}$`, owner["boot_id"]) || !match(`^time:\[[0-9]+\]$`, owner["namespace"]) {
		return nil, errArtifact
	}
	process := fmt.Sprintf("/proc/%.0f", owner["pid"])
	stat, err := os.ReadFile(process + "/stat")
	if err != nil {
		return nil, err
	}
	end := strings.LastIndex(string(stat), ") ")
	if end < 0 {
		return nil, errArtifact
	}
	fields := strings.Fields(string(stat)[end+2:])
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return nil, err
	}
	namespace, err := os.Readlink(process + "/ns/time")
	if err != nil {
		return nil, err
	}
	if len(fields) < 20 || fields[0] == "Z" || fields[0] == "X" || fields[19] != owner["start_ticks"] || strings.TrimSpace(string(boot)) != owner["boot_id"] || namespace != owner["namespace"] {
		return nil, errArtifact
	}
	return owner, nil
}
func wrapperGraphs(root string, owner object) ([]object, error) {
	directory, err := openDirectory(root, true)
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	names, err := directory.Readdirnames(67)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if len(names) > 66 {
		return nil, errArtifact
	}
	sort.Strings(names)
	graphs := []object{}
	ids := map[string]bool{}
	pattern := regexp.MustCompile(`^graph-[a-f0-9]{64}\.json$`)
	for _, name := range names {
		if name == "owner.json" || name == "completion.json" {
			continue
		}
		if !pattern.MatchString(name) {
			return nil, errArtifact
		}
		graph, _, err := boundObject(filepath.Join(root, name), "", 65536)
		if err != nil {
			return nil, err
		}
		if !exactKeys(graph, "version", "invocation_digest", "run_root", "run_id", "source_digest", "graph_digest", "phase", "scan_start_ms", "scan_end_ms", "scan_outcome") || graph["version"] != float64(1) || graph["invocation_digest"] != digest([]byte(owner["invocation_id"].(string))) || !match(`^[A-Za-z0-9_.-]{1,96}$`, graph["run_id"]) || !match(`^sha256:[a-f0-9]{64}$`, graph["source_digest"]) || !match(`^sha256:[a-f0-9]{64}$`, graph["graph_digest"]) {
			return nil, errArtifact
		}
		runRoot, ok := graph["run_root"].(string)
		if !ok || !filepath.IsAbs(runRoot) || filepath.Clean(runRoot) != runRoot || filepath.Base(runRoot) != graph["run_id"] || name != "graph-"+digest([]byte(runRoot))[7:]+".json" {
			return nil, errArtifact
		}
		phase, ok := graph["phase"].(string)
		if !ok || !contains([]string{"registered", "scanning", "scanned", "complete"}, phase) {
			return nil, errArtifact
		}
		if graph["scan_outcome"] != nil && graph["scan_outcome"] != "passed" && graph["scan_outcome"] != "failed" {
			return nil, errArtifact
		}
		for _, key := range []string{"scan_start_ms", "scan_end_ms"} {
			if graph[key] != nil && !safeInteger(graph[key], owner["start_ms"].(float64), 9007199254740991) {
				return nil, errArtifact
			}
		}
		id := graph["run_id"].(string)
		if ids[id] {
			return nil, errArtifact
		}
		ids[id] = true
		graphs = append(graphs, graph)
	}
	return graphs, nil
}

// PublishTerminalReceipts is the sole post-scan writer. All identity, schema,
// digest, and membership checks close before the first constrained publication.
func PublishTerminalReceipts(root string) error { return publishTerminalReceipts(root, publishNew) }
func publishTerminalReceipts(root string, publish func(string, []byte) error) error {
	owner, err := wrapperOwner(root)
	if err != nil {
		return err
	}
	graphs, err := wrapperGraphs(root, owner)
	if err != nil {
		return err
	}
	completion, _, err := boundObject(filepath.Join(root, "completion.json"), "", 65536)
	if err != nil {
		return err
	}
	start := owner["start_ms"].(float64)
	if !exactKeys(completion, "end_ms", "exit_code", "signal", "boot_id", "namespace") || !safeInteger(completion["end_ms"], start, 9007199254740991) || !safeInteger(completion["exit_code"], 0, 255) || (completion["signal"] != float64(0) && completion["signal"] != float64(2) && completion["signal"] != float64(15)) || completion["boot_id"] != owner["boot_id"] || completion["namespace"] != owner["namespace"] {
		return errArtifact
	}
	if len(graphs) == 0 {
		return nil
	}
	end := completion["end_ms"].(float64)
	source := graphs[0]["source_digest"]
	members := []any{}
	allPassed := true
	for _, graph := range graphs {
		if graph["phase"] != "complete" || graph["scan_outcome"] != "passed" || graph["scan_start_ms"] == nil || graph["scan_end_ms"] == nil || graph["scan_end_ms"].(float64) < graph["scan_start_ms"].(float64) || graph["scan_end_ms"].(float64) > end || graph["source_digest"] != source {
			return errArtifact
		}
		runRoot := graph["run_root"].(string)
		envelope, envelopeDigest, err := boundObject(filepath.Join(runRoot, "diagnostics/invocation-envelope.json"), "cartulary.harness_invocation_envelope.v2", 65536)
		if err != nil {
			return err
		}
		for _, key := range []string{"run_id", "source_digest", "graph_digest", "invocation_digest"} {
			if envelope[key] != graph[key] {
				return errArtifact
			}
		}
		previous := float64(0)
		for _, value := range envelope["intervals"].([]any) {
			interval := value.(object)
			if interval["start_ms"] != previous || interval["end_ms"].(float64) < previous {
				return errArtifact
			}
			previous = interval["end_ms"].(float64)
		}
		manifest, _, err := boundObject(filepath.Join(runRoot, "run-manifest.json"), "cartulary.harness_run_manifest.v3", 16*1024*1024)
		if err != nil {
			return err
		}
		for _, key := range []string{"run_id", "source_digest", "graph_digest"} {
			if manifest[key] != graph[key] {
				return errArtifact
			}
		}
		if manifest["instrumentation"].(object)["mode"] != "basic" {
			return errArtifact
		}
		summary, _, err := boundObject(filepath.Join(runRoot, "run-summary.json"), "cartulary.harness_run_summary.v1", 16*1024*1024)
		if err != nil {
			return err
		}
		if summary["run_id"] != graph["run_id"] || summary["target"] != manifest["target"] {
			return errArtifact
		}
		allPassed = allPassed && summary["status"] == "pass"
		_, scanDigest, err := boundObject(filepath.Join(runRoot, "retained-secret-scan.json"), "cartulary.harness_retained_secret_scan.v1", 65536)
		if err != nil {
			return err
		}
		members = append(members, object{"run_digest": digest([]byte(graph["run_id"].(string))), "graph_digest": graph["graph_digest"], "envelope_digest": envelopeDigest, "scan_digest": scanDigest, "scan_outcome": "passed", "scan_interval": object{"start_ms": graph["scan_start_ms"].(float64) - start, "end_ms": graph["scan_end_ms"].(float64) - start}})
	}
	now, err := bootMilliseconds()
	if err != nil {
		return err
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return err
	}
	namespace, err := os.Readlink("/proc/self/ns/time")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(boot)) != owner["boot_id"] || namespace != owner["namespace"] || now < end {
		return errArtifact
	}
	outcome := "failed"
	if completion["signal"] != float64(0) {
		outcome = "cancelled"
	} else if completion["exit_code"] == float64(0) && allPassed {
		outcome = "passed"
	}
	receipt := object{"schema_id": "cartulary.harness_terminal_receipt.v1", "invocation_digest": digest([]byte(owner["invocation_id"].(string))), "source_digest": source, "outer_duration_ms": end - start, "child_exit_code": completion["exit_code"], "outcome": outcome, "graph_memberships": members, "publication_start_ms": now - start}
	if err := Validate(receipt["schema_id"].(string), receipt); err != nil {
		return err
	}
	data, err := json.Marshal(receipt)
	if err != nil || len(data)+1 > 65536 {
		return errArtifact
	}
	data = append(data, '\n')
	roots := make([]string, 0, len(graphs))
	for _, graph := range graphs {
		roots = append(roots, graph["run_root"].(string))
	}
	return publishClosedReceipts(roots, data, publish)
}
func publishClosedReceipts(roots []string, data []byte, publish func(string, []byte) error) error {
	for _, root := range roots {
		if err := publish(filepath.Join(root, "diagnostics/terminal-receipt.json"), data); err != nil {
			return err
		}
	}
	return nil
}
