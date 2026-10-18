package harnessdiag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type object = map[string]any

var errUsage = errors.New("invalid observation inputs")

func validToken(value string, maximum int, initialUnderscore bool) bool {
	if len(value) == 0 || len(value) > maximum {
		return false
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if c == '_' && (i > 0 || initialUnderscore) || i > 0 && (c == '.' || c == '-') {
			continue
		}
		return false
	}
	return true
}

const observationSchema = "cartulary.harness_run_observation.v3"
const observationLimit = 64 * 1024

type Inputs struct{ ResultsDir, RunID, Target, AfterRevision, WaitSeconds string }
type selection struct {
	root, runID, target string
	after               *float64
	wait                time.Duration
}

func numberInput(text string, maximum uint64) (*float64, error) {
	if text == "" {
		return nil, nil
	}
	for _, c := range text {
		if c < '0' || c > '9' {
			return nil, errUsage
		}
	}
	n, err := strconv.ParseUint(text, 10, 64)
	if err != nil || n > maximum {
		return nil, errUsage
	}
	value := float64(n)
	return &value, nil
}
func normalize(in Inputs) (selection, error) {
	s := selection{}
	if in.ResultsDir == "" || len(in.ResultsDir) > 4096 || strings.ContainsAny(in.ResultsDir, "\\") {
		return s, errUsage
	}
	for _, c := range in.ResultsDir {
		if c < 32 {
			return s, errUsage
		}
	}
	for _, part := range strings.Split(in.ResultsDir, "/") {
		if part == ".." {
			return s, errUsage
		}
	}
	if in.RunID != "" && !validToken(in.RunID, 128, false) || in.Target != "" && !validToken(in.Target, 255, true) {
		return s, errUsage
	}
	root, err := filepath.Abs(filepath.Join(in.ResultsDir, in.RunID))
	if err != nil || len(root) > 4096 {
		return s, errUsage
	}
	id := filepath.Base(root)
	if !validToken(id, 128, false) {
		return s, errUsage
	}
	after, err := numberInput(in.AfterRevision, 9007199254740991)
	if err != nil {
		return s, err
	}
	wait, err := numberInput(in.WaitSeconds, 30)
	if err != nil {
		return s, err
	}
	duration := time.Duration(0)
	if wait != nil {
		duration = time.Duration(*wait) * time.Second
	}
	if duration > 0 && after == nil {
		return s, errUsage
	}
	return selection{root, id, in.Target, after, duration}, nil
}
func emptyObservation() object {
	return object{"schema_id": observationSchema, "operation_status": "ok", "operation_exit_code": float64(0), "run": nil, "selected_target": nil, "observation": "not_published", "revision": nil, "wait_result": "immediate", "waited_ms": float64(0), "live": nil, "terminal": nil, "unavailable_reason": nil, "limitations": []any{"current_source_not_audited", "outer_exit_not_observed", "process_not_observed"}, "failure_class": nil, "failure_reason": nil}
}
func addLimitation(result object, code string) {
	values := result["limitations"].([]any)
	for _, value := range values {
		if value == code {
			return
		}
	}
	values = append(values, code)
	sort.Slice(values, func(i, j int) bool { return values[i].(string) < values[j].(string) })
	result["limitations"] = values
}
func terminal(result, summary object, ref string) object {
	result["observation"] = "terminal_summary"
	result["wait_result"] = "terminal"
	result["terminal"] = object{"reported_status": summary["status"], "summary_ref": ref, "failure_class": summary["failure_class"], "failure_reason": summary["failure_reason"], "normalized_exit_code": summary["exit_code"], "evidence_audit": "not_run"}
	return result
}
func stringsOf(value any) []string {
	values := value.([]any)
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = v.(string)
	}
	return out
}
func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func sortedUnique(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1] >= values[i] {
			return false
		}
	}
	return true
}
func validateLive(snapshot, manifest object) error {
	if err := Validate("cartulary.harness_live_snapshot.v3", snapshot); err != nil {
		return err
	}
	for key, manifestKey := range map[string]string{"run_id": "run_id", "public_target": "target", "command_id": "command_id", "source_digest": "source_digest", "graph_digest": "graph_digest"} {
		if snapshot[key] != manifest[manifestKey] {
			return errArtifact
		}
	}
	if snapshot["availability"] == "unavailable" {
		return nil
	}
	ids := []string{}
	members := map[string]bool{}
	for _, value := range snapshot["units"].([]any) {
		id := value.(object)["unit_id"].(string)
		ids = append(ids, id)
		members[id] = true
	}
	if !sortedUnique(ids) {
		return errArtifact
	}
	targets := []string{}
	for _, value := range snapshot["targets"].([]any) {
		target := value.(object)
		targets = append(targets, target["target"].(string))
		ids := stringsOf(target["unit_ids"])
		if !sortedUnique(ids) {
			return errArtifact
		}
		for _, id := range ids {
			if !members[id] {
				return errArtifact
			}
		}
	}
	if !sortedUnique(targets) || !contains(targets, manifest["target"].(string)) {
		return errArtifact
	}
	for _, value := range snapshot["units"].([]any) {
		u := value.(object)
		resources := stringsOf(u["blocking_resources"])
		ids := stringsOf(u["blocking_unit_ids"])
		if !sortedUnique(resources) || !sortedUnique(ids) {
			return errArtifact
		}
		for _, id := range ids {
			if !members[id] {
				return errArtifact
			}
		}
		if u["started_elapsed_ms"] != nil && u["started_elapsed_ms"].(float64) > snapshot["sampled_elapsed_ms"].(float64) || u["failure_class"] != nil && u["failure_reason"] == nil {
			return errArtifact
		}
	}
	return nil
}
func capList(values []any, limit int) ([]any, float64) {
	n := len(values)
	if n > limit {
		return values[:limit], float64(n - limit)
	}
	return values, float64(0)
}
func projectLive(snapshot object, target string) (object, error) {
	units := snapshot["units"].([]any)
	if target != "" {
		var selected []string
		found := false
		for _, value := range snapshot["targets"].([]any) {
			v := value.(object)
			if v["target"] == target {
				selected = stringsOf(v["unit_ids"])
				found = true
				break
			}
		}
		if !found {
			return nil, errUsage
		}
		filtered := []any{}
		for _, v := range units {
			if contains(selected, v.(object)["unit_id"].(string)) {
				filtered = append(filtered, v)
			}
		}
		units = filtered
	}
	counts := object{"total": float64(len(units)), "pending": float64(0), "running": float64(0), "passed": float64(0), "failed": float64(0), "skipped": float64(0), "cancelled": float64(0)}
	caches := object{"not_observed": float64(0), "hit": float64(0), "miss": float64(0), "bypass": float64(0)}
	running, waiting, failures := []any{}, []any{}, []any{}
	for _, value := range units {
		u := value.(object)
		status := u["status"].(string)
		counts[status] = counts[status].(float64) + 1
		cache := u["cache_disposition"].(string)
		caches[cache] = caches[cache].(float64) + 1
		if status == "running" {
			var elapsed any
			if u["started_elapsed_ms"] != nil {
				elapsed = snapshot["sampled_elapsed_ms"].(float64) - u["started_elapsed_ms"].(float64)
			}
			running = append(running, object{"unit_id": u["unit_id"], "activity": u["activity"], "elapsed_ms": elapsed})
		}
		if u["wait_reason"] != nil {
			resources, ro := capList(u["blocking_resources"].([]any), 8)
			blockers, bo := capList(u["blocking_unit_ids"].([]any), 8)
			waiting = append(waiting, object{"unit_id": u["unit_id"], "activity": u["activity"], "wait_reason": u["wait_reason"], "blocking_resources": resources, "blocking_unit_ids": blockers, "blocking_resources_omitted": ro, "blocking_unit_ids_omitted": bo})
		}
		if u["failure_class"] != nil {
			failures = append(failures, object{"unit_id": u["unit_id"], "failure_class": u["failure_class"], "failure_reason": u["failure_reason"]})
		}
	}
	running, ro := capList(running, 16)
	waiting, wo := capList(waiting, 16)
	failures, fo := capList(failures, 16)
	live := object{"sampled_at": snapshot["sampled_at"], "sampled_elapsed_ms": snapshot["sampled_elapsed_ms"], "last_transition_seq": snapshot["last_transition_seq"], "publication_mode": snapshot["publication_mode"], "phase": snapshot["phase"], "unit_counts": counts, "cache_counts": caches, "running": running, "waiting": waiting, "failures_observed": failures, "running_omitted": ro, "waiting_omitted": wo, "failures_omitted": fo}
	if snapshot["resources"] != nil {
		live["resources"] = snapshot["resources"]
	}
	return live, nil
}
func sample(s selection) (object, error) {
	r := emptyObservation()
	identity := object{"result_root": s.root, "run_id": s.runID, "public_target": nil, "command_id": nil, "source_digest": nil, "graph_digest": nil}
	r["run"] = identity
	if s.target != "" {
		r["selected_target"] = s.target
	}
	directory, err := openDirectory(s.root, true)
	if errors.Is(err, os.ErrNotExist) {
		addLimitation(r, "live_unavailable")
		return r, nil
	}
	if err != nil {
		return r, errArtifact
	}
	defer directory.Close()
	manifest, err := readObject(filepath.Join(s.root, "run-manifest.json"), "cartulary.harness_run_manifest.v3", true, false)
	if err != nil {
		return r, err
	}
	if manifest != nil {
		if manifest["run_id"] != s.runID {
			return r, errArtifact
		}
		for _, key := range []string{"command_id", "source_digest", "graph_digest"} {
			identity[key] = manifest[key]
		}
		identity["public_target"] = manifest["target"]
	}
	summary, err := readObject(filepath.Join(s.root, "run-summary.json"), "cartulary.harness_run_summary.v1", true, false)
	if err != nil {
		return r, err
	}
	if summary != nil {
		if manifest == nil || summary["run_id"] != s.runID || summary["target"] != manifest["target"] {
			return r, errArtifact
		}
		if s.target == "" {
			return terminal(r, summary, "run-summary.json"), nil
		}
		ref := "target-summaries/" + s.target + ".json"
		if !contains(stringsOf(summary["artifact_refs"]), ref) {
			return r, errUsage
		}
		projection, err := readObject(filepath.Join(s.root, ref), "cartulary.harness_target_summary.v1", false, false)
		if err != nil {
			return r, err
		}
		if projection["target"] != s.target {
			return r, errArtifact
		}
		return terminal(r, projection, ref), nil
	}
	if manifest == nil {
		candidates := []string{s.target}
		if s.target == "" {
			candidates = []string{}
			entries, err := directory.ReadDir(4097)
			if (err != nil && !errors.Is(err, io.EOF)) || len(entries) > 4096 {
				return r, errArtifact
			}
			for _, entry := range entries {
				if entry.IsDir() && validToken(entry.Name(), 255, true) {
					candidates = append(candidates, entry.Name())
				}
			}
			sort.Strings(candidates)
		}
		var selected object
		ref := ""
		for _, candidate := range candidates {
			candidateRef := candidate + "/tool-run-summary.json"
			tool, err := readObject(filepath.Join(s.root, candidateRef), "cartulary.tool_run_summary.v5", true, false)
			if err != nil {
				return r, err
			}
			if tool == nil {
				continue
			}
			runRoot, e1 := filepath.Abs(tool["run_root"].(string))
			resultRoot, e2 := filepath.Abs(tool["result_root"].(string))
			if e1 != nil || e2 != nil || tool["run_id"] != s.runID || tool["target"] != candidate || runRoot != s.root || resultRoot != filepath.Dir(s.root) {
				return r, errArtifact
			}
			if selected != nil {
				return r, errUsage
			}
			selected = tool
			ref = candidateRef
		}
		if selected != nil {
			identity["public_target"] = selected["target"]
			return terminal(r, selected, ref), nil
		}
		addLimitation(r, "live_unavailable")
		return r, nil
	}
	snapshot, err := readObject(filepath.Join(s.root, "diagnostics/live-status.json"), "", true, true)
	if err != nil {
		return r, err
	}
	if snapshot == nil {
		addLimitation(r, "live_unavailable")
		return r, nil
	}
	if err := validateLive(snapshot, manifest); err != nil {
		return r, err
	}
	revision := snapshot["revision"].(float64)
	if s.after != nil && *s.after > revision {
		return r, errUsage
	}
	r["revision"] = revision
	if snapshot["publication_mode"] == "paused_measurement" {
		r["wait_result"] = "unavailable"
		addLimitation(r, "measurement_pause")
	}
	if snapshot["availability"] == "unavailable" {
		r["observation"] = "unavailable"
		r["unavailable_reason"] = snapshot["unavailable_reason"]
		r["wait_result"] = "unavailable"
		addLimitation(r, "live_unavailable")
	} else {
		live, err := projectLive(snapshot, s.target)
		if err != nil {
			return r, err
		}
		r["observation"] = "live"
		r["live"] = live
	}
	return r, nil
}
func failObservation(r object, err error, code int) object {
	class, reason := "artifact", "artifact_error"
	if errors.Is(err, errUsage) {
		class, reason, code = "config", "usage_error", 2
	} else if errors.Is(err, context.Canceled) {
		class, reason = "interrupted", "cancelled_or_interrupted"
		if code == 0 {
			code = 15
		}
	} else {
		code = 11
	}
	r["operation_status"] = "error"
	r["operation_exit_code"] = float64(code)
	r["observation"] = "invalid"
	r["live"] = nil
	r["terminal"] = nil
	r["failure_class"] = class
	r["failure_reason"] = reason
	return r
}
func bound(r object) (object, error) {
	if live, ok := r["live"].(object); ok {
		lists := [][2]string{{"failures_observed", "failures_omitted"}, {"waiting", "waiting_omitted"}, {"running", "running_omitted"}}
		for _, pair := range lists {
			if live[pair[1]].(float64) > 0 {
				addLimitation(r, "lists_truncated")
			}
		}
		for _, value := range live["waiting"].([]any) {
			v := value.(object)
			if v["blocking_resources_omitted"].(float64) > 0 || v["blocking_unit_ids_omitted"].(float64) > 0 {
				addLimitation(r, "lists_truncated")
			}
		}
		for {
			data, err := json.Marshal(r)
			if err != nil {
				return r, errArtifact
			}
			if len(data)+1 <= observationLimit {
				break
			}
			removed := false
			for _, pair := range lists {
				list := live[pair[0]].([]any)
				if len(list) > 0 {
					live[pair[0]] = list[:len(list)-1]
					live[pair[1]] = live[pair[1]].(float64) + 1
					removed = true
					break
				}
			}
			if !removed {
				return r, errArtifact
			}
			addLimitation(r, "lists_truncated")
		}
	}
	if err := Validate(observationSchema, r); err != nil {
		return r, err
	}
	data, err := json.Marshal(r)
	if err != nil || len(data)+1 > observationLimit {
		return r, errArtifact
	}
	return r, nil
}

// Observe reads only selected public evidence. The context cancels this reader,
// never the invocation. It does not establish liveness or audit acceptance.
func Observe(ctx context.Context, in Inputs) object {
	start := time.Now()
	r := emptyObservation()
	s, err := normalize(in)
	if err == nil {
		for {
			if err = ctx.Err(); err != nil {
				break
			}
			next, sampleErr := sample(s)
			if sampleErr != nil {
				err = sampleErr
				break
			}
			r = next
			r["waited_ms"] = float64(time.Since(start).Milliseconds())
			if r["wait_result"] == "terminal" || r["wait_result"] == "unavailable" {
				break
			}
			if s.after != nil && r["revision"] != nil && r["revision"].(float64) > *s.after {
				r["wait_result"] = "changed"
				break
			}
			if s.wait == 0 {
				break
			}
			remaining := s.wait - time.Since(start)
			if remaining <= 0 {
				r["wait_result"] = "wait_expired"
				break
			}
			if remaining > time.Second {
				remaining = time.Second
			}
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(remaining):
			}
			if err != nil {
				break
			}
			if time.Since(start) >= s.wait {
				r["wait_result"] = "wait_expired"
				r["waited_ms"] = float64(time.Since(start).Milliseconds())
				break
			}
		}
	}
	if err != nil {
		r = failObservation(r, err, 0)
		r["waited_ms"] = float64(time.Since(start).Milliseconds())
	}
	if _, err = bound(r); err != nil {
		r = failObservation(r, err, 0)
	}
	return r
}
func Render(r object) string {
	run, _ := r["run"].(object)
	root, target := "unavailable", "unavailable"
	if run != nil {
		root = run["result_root"].(string)
		if run["public_target"] != nil {
			target = run["public_target"].(string)
		}
	}
	if r["selected_target"] != nil {
		target = r["selected_target"].(string)
	}
	lines := []string{fmt.Sprintf("Observation: %v; reader=%v code=%v", r["observation"], r["operation_status"], r["operation_exit_code"]), "Run: " + root, fmt.Sprintf("Target: %s; wait=%v", target, r["wait_result"])}
	if live, ok := r["live"].(object); ok {
		lines = append(lines, fmt.Sprintf("Sample: %v; revision=%v; progress=%v; %v; %v", live["sampled_at"], r["revision"], live["last_transition_seq"], live["publication_mode"], live["phase"]))
		counts, _ := json.Marshal(live["unit_counts"])
		lines = append(lines, "Units: "+string(counts))
		for _, pair := range [][2]string{{"running", "Running"}, {"waiting", "Waiting"}, {"failures_observed", "Failure"}} {
			list := live[pair[0]].([]any)
			for _, v := range list[:min(len(list), 4)] {
				u := v.(object)
				detail := fmt.Sprint(u["activity"])
				if pair[0] == "waiting" {
					detail = fmt.Sprint(u["wait_reason"])
				}
				if pair[0] == "failures_observed" {
					lines = append(lines, fmt.Sprintf("Failure: %v %v/%v", u["unit_id"], u["failure_class"], u["failure_reason"]))
				} else {
					lines = append(lines, fmt.Sprintf("%s: %v (%s)", pair[1], u["unit_id"], detail))
				}
			}
		}
		lines = append(lines, fmt.Sprintf("Omitted: running=%v waiting=%v failures=%v", live["running_omitted"].(float64)+float64(max(0, len(live["running"].([]any))-4)), live["waiting_omitted"].(float64)+float64(max(0, len(live["waiting"].([]any))-4)), live["failures_omitted"].(float64)+float64(max(0, len(live["failures_observed"].([]any))-4))))
	}
	if terminal, ok := r["terminal"].(object); ok {
		lines = append(lines, fmt.Sprintf("Reported result: %v; %v; evidence audit not run", terminal["reported_status"], terminal["summary_ref"]))
	}
	if r["failure_class"] != nil {
		lines = append(lines, fmt.Sprintf("Reader failure: %v/%v", r["failure_class"], r["failure_reason"]))
	}
	lines = append(lines, "Limitations: "+strings.Join(stringsOf(r["limitations"]), ", "))
	output := strings.Builder{}
	for i, line := range lines {
		if i >= 29 || output.Len()+len(line)+1 > 8100 {
			output.WriteString("Display truncated; use JSON=1 for bounded structured detail.\n")
			break
		}
		output.WriteString(line + "\n")
	}
	return output.String()
}
