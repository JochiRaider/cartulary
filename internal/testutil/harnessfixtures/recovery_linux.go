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

var ErrCleanupBlocked = errors.New("resource cleanup blocked by unresolved dependent")

type CleanupOutcome struct {
	Kind, Outcome string
	Err           error
}

func (p ProductionOptions) stopResource(ctx context.Context, resource harnesssession.Resource) error {
	return CleanupFailure(p.Runtime.StopResource(ctx, p.Root, p.RunRoot, p.environment(), resource))
}
func (p ProductionOptions) releaseResource(resource harnesssession.Resource) error {
	return p.Runtime.RecordResource(resource.Kind, resource.Target, "released")
}

// Recover uses only this runtime's verified private records. Unknown browser
// dependency proof blocks suite termination. Borrowed suites are detached only.
func (p ProductionOptions) Recover(ctx context.Context) error {
	p.RecoveryOutput = true
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
	unknownDependent := false
	report := func(resource harnesssession.Resource, err error) {
		outcome := "completed"
		if err != nil {
			outcome = "failed"
			failures = append(failures, err)
		}
		if errors.Is(err, ErrCleanupBlocked) {
			outcome = "blocked"
		}
		if p.ReportCleanup != nil {
			p.ReportCleanup(CleanupOutcome{resource.Kind, outcome, err})
		}
	}
	block := func(resource harnesssession.Resource, fallback string) {
		dependencies := resource.Dependencies
		if dependencies == nil && fallback != "" {
			dependencies = []string{fallback}
		}
		if dependencies == nil {
			unknownDependent = true
			return
		}
		for _, dependency := range dependencies {
			blocked[dependency] = true
		}
	}
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
			err := p.stopResource(ctx, resource)
			if err == nil {
				err = p.releaseResource(resource)
			}
			report(resource, err)
			if err != nil {
				block(resource, "")
			}
		}
	}
	for _, resource := range resources {
		if resource.Kind != "browser_stack" {
			continue
		}
		file, _ := resource.Path()
		if unknownDependent || blocked[file] {
			report(resource, ErrCleanupBlocked)
			block(resource, "")
			continue
		}
		dependency := ""
		err := p.Runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error {
			consumer := p
			consumer.Dependencies = append([]string{file}, resource.Dependencies...)
			owner, err := consumer.browserOperation(step, "inspect", []string{file}, p.environment())
			var known bool
			dependency, known = owner["suite_lease"].(string)
			if err != nil || !known {
				dependency = ""
				return artifactFailure(errors.New("browser dependency unavailable"))
			}
			if resource.Dependencies != nil && (len(resource.Dependencies) != 1 || resource.Dependencies[0] != dependency) {
				unknownDependent = true
				return artifactFailure(errors.New("browser dependency contradicts acquisition intent"))
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
			output, err := p.Runtime.AllocateRecoveryOutput()
			if err != nil {
				return err
			}
			bound["CARTULARY_BROWSER_RECOVERY_RESULTS_DIR"] = filepath.Dir(output)
			consumer.Dependencies = []string{file, dependency}
			_, err = consumer.browserOperation(step, "recover", []string{file}, bound)
			return err
		})
		if err == nil {
			err = p.releaseResource(resource)
		}
		if err != nil {
			block(resource, dependency)
		}
		report(resource, err)

	}
	for _, resource := range resources {
		if resource.Kind != "object_store_proxy" {
			continue
		}
		file, _ := resource.Path()
		if unknownDependent || blocked[file] {
			report(resource, ErrCleanupBlocked)
			block(resource, "")
			continue
		}
		err := p.stopResource(ctx, resource)
		if err == nil {
			err = p.releaseResource(resource)
		}
		report(resource, err)
		if err != nil {
			block(resource, "")
		}
	}
	for _, resource := range resources {
		if resource.Kind == "borrowed_suite" {
			if unknownDependent || blocked[borrowedDependency] {
				report(resource, ErrCleanupBlocked)
				continue
			}
			_, err := p.attachmentOperation(ctx, "detach")
			if err == nil {
				err = p.releaseResource(resource)
			}
			report(resource, err)
		}
	}
	for _, file := range slices.Sorted(func(yield func(string) bool) {
		for key := range suites {
			if !yield(key) {
				return
			}
		}
	}) {
		if unknownDependent || blocked[file] {
			report(suites[file], ErrCleanupBlocked)
			continue
		}
		environment, err := readRecoveryEnvironment(filepath.Join(filepath.Dir(file), "suite-environment.json"))
		if err != nil {
			report(suites[file], artifactFailure(err))
			continue
		}
		bound := p.environment()
		maps.Copy(bound, environment)
		report(suites[file], p.stopSuite(ctx, file, bound))
	}
	return errors.Join(failures...)
}
