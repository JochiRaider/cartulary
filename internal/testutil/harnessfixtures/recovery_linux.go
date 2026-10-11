//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func (p ProductionOptions) stopResource(ctx context.Context, kind string, resource Resource) error {
	_, err := p.command()(ctx, p.Environment["NODE_BIN"], []string{filepath.Join(p.Root, "tools/harness/runtime/fixture-resource-stop.mjs"), kind, string(resource.Target)}, p.environment())
	return CleanupFailure(err)
}
func (p ProductionOptions) releaseResource(resource Resource) error {
	var target any
	if err := json.Unmarshal(resource.Target, &target); err != nil {
		return err
	}
	return p.Runtime.RecordResource(resource.Kind, target, "released")
}

// Recover uses only this runtime's verified private records. Unknown browser
// dependency proof blocks suite termination. Borrowed suites are detached only.
func (p ProductionOptions) Recover(ctx context.Context) error {
	resources, err := p.Runtime.Resources(4096)
	if err != nil {
		return artifactFailure(err)
	}
	failures := []error{}
	blocked := map[string]bool{}
	suites := map[string]Resource{}
	for _, resource := range resources {
		if resource.Kind == "managed_suite" {
			path, err := resource.Path()
			if err != nil {
				return err
			}
			suites[path] = resource
		}
	}
	borrowedDependency := filepath.Join(p.Runtime.Root, "borrowed-service-lease.json")
	borrowed := false
	for _, r := range resources {
		borrowed = borrowed || r.Kind == "borrowed_suite"
	}
	allBlocked := false
	for _, resource := range resources {
		operation := ""
		if resource.Kind == "browser_renderer" {
			operation = "renderer"
		} else if slices.Contains([]string{"helper_process", "browser_process"}, resource.Kind) {
			var proof processTarget
			if json.Unmarshal(resource.Target, &proof) != nil {
				return harnesslocal.ErrUnsafe
			}
			if proof.PID == os.Getpid() {
				continue
			}
			operation = "process"
		}
		if operation != "" {
			if err := p.stopResource(ctx, operation, resource); err != nil {
				failures = append(failures, err)
				allBlocked = true
			} else if err := p.releaseResource(resource); err != nil {
				failures = append(failures, err)
				allBlocked = true
			}
		}
	}
	for _, resource := range resources {
		if resource.Kind != "browser_stack" {
			continue
		}
		file, _ := resource.Path()
		owner, err := p.browserOperation(ctx, "inspect", []string{file}, p.environment())
		dependency, known := owner["suite_lease"].(string)
		if err != nil || !known {
			failures = append(failures, artifactFailure(errors.New("browser dependency unavailable")))
			allBlocked = true
			continue
		}
		if _, ok := suites[dependency]; !ok && !(borrowed && dependency == borrowedDependency) {
			failures = append(failures, artifactFailure(errors.New("browser dependency not owned by this recovery")))
			allBlocked = true
			continue
		}
		if _, err := p.browserOperation(ctx, "recover", []string{file}, p.environment()); err != nil {
			failures = append(failures, err)
			blocked[dependency] = true
			continue
		}
		if err := p.releaseResource(resource); err != nil {
			failures = append(failures, err)
			blocked[dependency] = true
		}
	}
	for _, resource := range resources {
		if resource.Kind != "object_store_proxy" {
			continue
		}
		if err := p.stopResource(ctx, "proxy", resource); err != nil {
			failures = append(failures, err)
			allBlocked = true
		} else if err := p.releaseResource(resource); err != nil {
			failures = append(failures, err)
			allBlocked = true
		}
	}
	for _, resource := range resources {
		if resource.Kind == "borrowed_suite" && !allBlocked && !blocked[borrowedDependency] {
			if _, err := p.attachmentOperation(ctx, "detach"); err != nil {
				failures = append(failures, err)
			} else if err := p.releaseResource(resource); err != nil {
				failures = append(failures, err)
			}
		}
	}
	for _, file := range slices.Sorted(func(yield func(string) bool) {
		for key := range suites {
			if !yield(key) {
				return
			}
		}
	}) {
		if allBlocked || blocked[file] {
			continue
		}
		if err := p.stopSuite(ctx, file, p.environment()); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
