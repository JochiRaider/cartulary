package referenceassembly

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type lifecycleRegistrar interface {
	RegisterHandler(string, jobs.HandlerFunc) error
}

// NewAdministration composes the complete owner and binds its single lifecycle
// worker independently of HTTP. Construction errors precede route publication.
func NewAdministration(options reference_data.CoordinatorOptions, runner lifecycleRegistrar) (reference_data.AdministrativeApplication, error) {
	if runner == nil {
		return nil, errors.New("reference pack: lifecycle runner required")
	}
	application, err := reference_data.NewCoordinator(options)
	if err != nil {
		return nil, err
	}
	if err := runner.RegisterHandler(reference_data.LifecycleWorkerKind, application.Execute); err != nil {
		return nil, err
	}
	return application, nil
}
