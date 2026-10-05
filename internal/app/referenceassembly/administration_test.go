package referenceassembly_test

import (
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// No operation is executed here; live Jobs tests exercise the constructed ports.
type constructionStorage struct{ reference_data.ArtifactStorage }
type constructionFinalizer struct {
	reference_data.JobSuccessFinalizer
}
type constructionRunner struct{ registered bool }

func (r *constructionRunner) RegisterHandler(kind string, handler jobs.HandlerFunc) error {
	if r.registered {
		return errors.New("duplicate worker")
	}
	if kind != reference_data.LifecycleWorkerKind || handler == nil {
		return errors.New("wrong worker")
	}
	r.registered = true
	return nil
}

func TestReferencePackAdministrationConstruction_Unit(t *testing.T) {
	valid := reference_data.CoordinatorOptions{
		Postgres: &pgxpool.Pool{}, Storage: constructionStorage{}, Limits: reference_data.DefaultLimits(),
		JobAdmission: &jobs.TransactionService{}, JobOperations: &jobs.Manager{}, JobExecutionGuard: &jobs.TransactionService{},
		RegistryUsage: referenceassembly.RegistryUsage{}, JobFinalizer: constructionFinalizer{}, Now: time.Now,
	}
	missing := []struct {
		name string
		omit func(*reference_data.CoordinatorOptions)
	}{
		{"database", func(o *reference_data.CoordinatorOptions) { o.Postgres = nil }},
		{"storage", func(o *reference_data.CoordinatorOptions) { o.Storage = nil }},
		{"admission", func(o *reference_data.CoordinatorOptions) { o.JobAdmission = nil }},
		{"operations", func(o *reference_data.CoordinatorOptions) { o.JobOperations = nil }},
		{"guard", func(o *reference_data.CoordinatorOptions) { o.JobExecutionGuard = nil }},
		{"finalizer", func(o *reference_data.CoordinatorOptions) { o.JobFinalizer = nil }},
		{"registry usage", func(o *reference_data.CoordinatorOptions) { o.RegistryUsage = nil }},
		{"clock", func(o *reference_data.CoordinatorOptions) { o.Now = nil }},
		{"limits", func(o *reference_data.CoordinatorOptions) { o.Limits.ReferencePacks.MaxVerificationSeconds = 0 }},
	}
	for _, test := range missing {
		t.Run(test.name, func(t *testing.T) {
			options := valid
			test.omit(&options)
			runner := &constructionRunner{}
			if app, err := referenceassembly.NewAdministration(options, runner); err == nil || app != nil || runner.registered {
				t.Fatal("incomplete application registered", app, err)
			}
		})
	}
	if app, err := referenceassembly.NewAdministration(valid, nil); err == nil || app != nil {
		t.Fatal("missing runner admitted", err)
	}
	runner := &constructionRunner{}
	if app, err := referenceassembly.NewAdministration(valid, runner); err != nil || app == nil || !runner.registered {
		t.Fatal("route-free construction failed", err)
	}
	if app, err := referenceassembly.NewAdministration(valid, runner); err == nil || app != nil {
		t.Fatal("duplicate registration admitted", err)
	}
	if admission, err := reference_data.NewImportAdmission(reference_data.ImportAdmissionOptions{Postgres: valid.Postgres, Storage: valid.Storage, Limits: valid.Limits, JobAdmission: valid.JobAdmission, Now: valid.Now}); err != nil || admission == nil {
		t.Fatal("admission-only construction requires executor", err)
	}
}
