package referenceassembly

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewOperatorAdmission composes only import admission and observation. Durable
// Jobs execution remains with the deployment's admitted lifecycle worker; this
// command creates neither a listener nor a second verification executor.
func NewOperatorAdmission(ctx context.Context, pool *pgxpool.Pool, storage reference_data.ArtifactStorage, configuration reference_data.Configuration, limits reference_data.Limits, requestedClaims []string, now func() time.Time) (reference_data.ImportAdmission, *jobs.Manager, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, nil, err
	}
	resolved, err := coordinator.ResolveClaims(requestedClaims)
	if err != nil {
		return nil, nil, err
	}
	plan, err := coordinator.BuildPublicationPlan(resolved)
	if err != nil {
		return nil, nil, err
	}
	publication, err := extensionassembly.NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		return nil, nil, err
	}
	if _, admitted := publication.Job("reference_pack.import_v2"); !admitted {
		return nil, nil, errors.New("reference pack: operator import profile is not admitted")
	}
	definitions, err := extensionassembly.RecognizedJobDefinitions(coordinator.JobKindContracts(), coordinator.WorkerRuntimeContracts())
	if err != nil {
		return nil, nil, err
	}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		return nil, nil, err
	}
	admitted, err := extensionassembly.JobDefinitions(publication)
	if err != nil {
		return nil, nil, err
	}
	workers, err := extensionassembly.WorkerRuntimeContracts(publication)
	if err != nil {
		return nil, nil, err
	}
	selection, err := jobs.NewRuntimeSelection(catalog, extensionassembly.JobKinds(admitted), workers)
	if err != nil {
		return nil, nil, err
	}
	ports := extensionassembly.JobOwnerTransactionAdapters{}
	transactions, err := jobs.NewTransactionService(deploymentAdmissionProgress{}, jobs.OwnerTransactionPorts{RouteIdempotency: ports, ExtensionCancellation: ports, TerminalEffects: reference_data.JobTerminalEffects{}}, catalog, selection)
	if err != nil {
		return nil, nil, err
	}
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: jobs.ProductionRuntimePolicy(), Now: now})
	if err != nil {
		return nil, nil, err
	}
	if err := manager.ValidateStorageCatalog(ctx); err != nil {
		return nil, nil, err
	}
	admission, err := reference_data.NewImportAdmission(reference_data.ImportAdmissionOptions{Postgres: pool, Storage: storage, Configuration: configuration, Limits: limits, JobAdmission: transactions, Now: now})
	if err != nil {
		return nil, nil, err
	}
	return admission, manager, nil
}

// Deployment-scoped Job admissions never emit incident collaboration intents.
// Reject a composition error rather than silently dropping an incident event.
type deploymentAdmissionProgress struct{}

func (deploymentAdmissionProgress) AppendProgressIntentTx(context.Context, pgx.Tx, jobs.ProgressIntent) error {
	return errors.New("reference pack: incident progress is outside local import admission")
}
