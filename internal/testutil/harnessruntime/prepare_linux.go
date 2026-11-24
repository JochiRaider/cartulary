//go:build linux

package harnessruntime

import (
	"fmt"
	"os"
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
