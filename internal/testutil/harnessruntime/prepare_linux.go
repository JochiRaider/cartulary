//go:build linux

package harnessruntime

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
)

// prepare plans the invocation with the Go planner inside the runtime. The
// runtime's own instrumentation mode, and the cache mode an accounted run
// imposes, take precedence over the command's inputs. It performs no unit,
// fixture, retained-publication or cache transaction.
func prepare(root string, args []string, environment map[string]string, mode, imposedCacheMode string) (*Plan, error) {
	request, err := planRequest(args, environment)
	if err != nil {
		return nil, plannerFailure(fmt.Errorf("%w: %v", harnessplan.ErrSelection, err))
	}
	request.InstrumentationMode, request.CacheMode = mode, imposedCacheMode
	source, err := harnessplan.Prepare(root, request, time.Now(), os.Getpid())
	if err != nil {
		return nil, plannerFailure(err)
	}
	return adoptPlan(source)
}

// PlanInvocation plans the invocation the runtime would execute for graph
// arguments args under environment, without executing or publishing
// anything. A rejection carries the closed preparation diagnostic.
func PlanInvocation(root string, args, environment []string) (*harnessplan.Plan, error) {
	env := make(map[string]string, len(environment))
	for _, entry := range environment {
		if name, value, ok := strings.Cut(entry, "="); ok {
			env[name] = value
		}
	}
	request, err := planRequest(args, env)
	if err != nil {
		return nil, plannerFailure(fmt.Errorf("%w: %v", harnessplan.ErrSelection, err))
	}
	plan, err := harnessplan.Prepare(root, request, time.Now(), os.Getpid())
	if err != nil {
		return nil, plannerFailure(err)
	}
	return plan, nil
}
