//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func (p ProductionOptions) stopResource(ctx context.Context, resource harnesssession.Resource) error {
	return CleanupFailure(p.Runtime.StopResource(ctx, p.Root, p.RunRoot, p.environment(), resource))
}
func (p ProductionOptions) releaseResource(resource harnesssession.Resource) error {
	return p.Runtime.RecordResource(resource.Kind, resource.Target, "released")
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
	suites := map[string]harnesssession.Resource{}
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
			if err := p.stopResource(ctx, resource); err != nil {
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
		dependency := ""
		err := p.Runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error {
			owner, err := p.browserOperation(step, "inspect", []string{file}, p.environment())
			var known bool
			dependency, known = owner["suite_lease"].(string)
			if err != nil || !known {
				dependency = ""
				return artifactFailure(errors.New("browser dependency unavailable"))
			}
			if _, ok := suites[dependency]; !ok && !(borrowed && dependency == borrowedDependency) {
				dependency = ""
				return artifactFailure(errors.New("browser dependency not owned by this recovery"))
			}
			environmentFile := filepath.Join(filepath.Dir(dependency), "suite-environment.json")
			if dependency == borrowedDependency {
				environmentFile = filepath.Join(p.Runtime.Root, "borrowed-environment.json")
			}
			environment, err := readRecoveryEnvironment(environmentFile)
			if err != nil {
				return artifactFailure(err)
			}
			bound := p.environment()
			maps.Copy(bound, environment)
			output, err := p.Runtime.PrivatePath("lifecycle", "browser-recovery-"+identifier(), p.Runtime.RunID)
			if err != nil {
				return err
			}
			if err = p.Runtime.AuthorizeOutputRoot(output); err != nil {
				return err
			}
			bound["CARTULARY_BROWSER_RECOVERY_RESULTS_DIR"] = filepath.Dir(output)
			_, err = p.browserOperation(step, "recover", []string{file}, bound)
			return err
		})
		if err == nil {
			err = p.releaseResource(resource)
		}
		if err != nil {
			failures = append(failures, err)
			if dependency == "" {
				allBlocked = true
			} else {
				blocked[dependency] = true
			}
		}

	}
	for _, resource := range resources {
		if resource.Kind != "object_store_proxy" {
			continue
		}
		if err := p.stopResource(ctx, resource); err != nil {
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
		environment, err := readRecoveryEnvironment(filepath.Join(filepath.Dir(file), "suite-environment.json"))
		if err != nil {
			failures = append(failures, artifactFailure(err))
			continue
		}
		bound := p.environment()
		maps.Copy(bound, environment)
		if err := p.stopSuite(ctx, file, bound); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
