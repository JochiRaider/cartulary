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
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

var ErrCleanupBlocked = errors.New("resource cleanup blocked by unresolved dependent")

type CleanupOutcome struct {
	Kind, Outcome string
	Err           error
}

func (p ProductionOptions) stopResource(ctx context.Context, resource harnesssession.Resource) error {
	// Resource identity is bound to the original run. Fixture output directories
	// are separate capabilities and may never have existed in dev/artifact mode.
	return CleanupFailure(p.Runtime.StopResource(ctx, p.Root, p.Runtime.RunRoot, p.environment(), resource))
}
func (p ProductionOptions) releaseResource(resource harnesssession.Resource) error {
	return p.Runtime.RecordResource(resource.Kind, resource.Target, "released")
}

// Recover and ordinary closure share one evaluator. Only physical operations
// differ; neither operation infers dependency authority from preservation paths.
func (p ProductionOptions) Recover(ctx context.Context) (result error) {
	if !p.Runtime.InRecoveryAttempt() {
		if _, err := p.Runtime.BeginRecoveryAttempt(); err != nil {
			return err
		}
		defer func() { result = errors.Join(result, p.Runtime.FinishRecoveryAttempt()) }()
	}
	p.RecoveryOutput = true
	var err error
	p.RunRoot, err = p.Runtime.RecoveryDetailPath("lifecycle", p.Runtime.RunID)
	if err != nil {
		return err
	}
	if err := p.Runtime.AuthorizeOutputRoot(p.RunRoot); err != nil {
		return err
	}
	return p.withResidentActivity(ctx, func(ctx context.Context) error { return p.closeResources(ctx, "", nil) })
}

func (p ProductionOptions) cleanupResource(ctx context.Context, resource harnesssession.Resource, resources []harnesssession.Resource) error {
	if strings.HasSuffix(resource.Kind, "_process") {
		var proof processTarget
		if json.Unmarshal(resource.Target, &proof) != nil {
			return harnesslocal.ErrUnsafe
		}
		if proof.PID == os.Getpid() {
			return ErrCleanupBlocked
		}
	}
	switch resource.Kind {
	case "resident_allocation":
		file, err := resource.Path()
		if err != nil {
			return err
		}
		return releaseResident(ctx, file)
	case "helper_process", "browser_process", "diagnostic_scope", "browser_renderer", "object_store_proxy":
		return p.stopResource(ctx, resource)
	case "borrowed_suite":
		p.Dependencies = []string{resource.Reference()}
		_, err := p.attachmentOperation(ctx, "detach")
		return err
	case "managed_suite":
		file, err := resource.Path()
		if err != nil {
			return err
		}
		environment, err := readRecoveryEnvironment(filepath.Join(filepath.Dir(file), "suite-environment.json"))
		if err != nil {
			return artifactFailure(err)
		}
		bound := p.environment()
		maps.Copy(bound, environment)
		return p.stopSuitePhysical(ctx, file, bound)
	case "browser_stack":
		return p.recoverBrowser(ctx, resource, resources)
	default:
		return errCleanupUncertain
	}
}

func (p ProductionOptions) recoverBrowser(ctx context.Context, resource harnesssession.Resource, resources []harnesssession.Resource) error {
	file, err := resource.Path()
	if err != nil {
		return err
	}
	return p.Runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error {
		if len(resource.Dependencies) < 1 || len(resource.Dependencies) > 2 {
			return errCleanupUncertain
		}
		var dependency *harnesssession.Resource
		for i := range resources {
			if (resources[i].Kind == "managed_suite" || resources[i].Kind == "borrowed_suite") && slices.Contains(resource.Dependencies, resources[i].Reference()) {
				dependency = &resources[i]
				break
			}
		}
		if dependency == nil || (dependency.Kind != "managed_suite" && dependency.Kind != "borrowed_suite") {
			return errCleanupUncertain
		}
		suite, err := dependency.Path()
		if err != nil {
			return err
		}
		environmentFile := filepath.Join(filepath.Dir(suite), "suite-environment.json")
		if dependency.Kind == "borrowed_suite" {
			suite = filepath.Join(p.Runtime.Root, "borrowed-service-lease.json")
			environmentFile = filepath.Join(p.Runtime.Root, "borrowed-environment.json")
		}
		p.Dependencies = append([]string{resource.Reference()}, resource.Dependencies...)
		owner, err := p.browserOperation(step, "inspect", []string{file}, p.environment())
		if err != nil {
			return err
		}
		if owner["suite_lease"] != suite {
			return errCleanupUncertain
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
		_, err = p.browserOperation(step, "recover", []string{file}, bound)
		return err
	})
}
