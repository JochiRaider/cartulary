package evidence

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records/subtypepresence"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
)

func TestModuleEvidenceProviderContributionClosure(t *testing.T) {
	t.Parallel()

	t.Run("incident_bundle", func(t *testing.T) {
		t.Parallel()
		descriptor := NewIncidentBundleSourcePort().Descriptor()
		if descriptor.FamilyID != "evidence" ||
			descriptor.ContractMajor != 2 ||
			descriptor.OwnerID != "module.evidence" {
			t.Fatalf("unexpected Evidence incident-bundle descriptor: %#v", descriptor)
		}
		wantPaths := []string{
			"data/evidence_records.ndjson",
			"data/evidence_custody_events.ndjson",
			"data/object_blobs.ndjson",
		}
		gotPaths := make([]string, 0, len(descriptor.Paths))
		for _, path := range descriptor.Paths {
			gotPaths = append(gotPaths, path.LogicalPath)
		}
		if !slices.Equal(gotPaths, wantPaths) {
			t.Fatalf("Evidence incident-bundle paths = %v, want %v", gotPaths, wantPaths)
		}
	})

	t.Run("recovery", func(t *testing.T) {
		t.Parallel()
		contribution := RecoveryStateContribution()
		if contribution.SchemaID != recoverystate.ContributionSchemaID ||
			contribution.OwnerID != "module.evidence" {
			t.Fatalf("unexpected Evidence recovery contribution: %#v", contribution)
		}
		wantTables := []string{
			"evidence",
			"evidence_custody_events",
			"object_blobs",
			"evidence_access_handles",
			"evidence_object_upload_leases",
			"evidence_blob_cleanup_claims",
		}
		gotTables := make([]string, 0, len(contribution.Tables))
		for _, table := range contribution.Tables {
			gotTables = append(gotTables, table.TableName)
		}
		if !slices.Equal(gotTables, wantTables) {
			t.Fatalf("Evidence recovery tables = %v, want %v", gotTables, wantTables)
		}
		if len(contribution.ObjectFamilies) != 1 ||
			contribution.ObjectFamilies[0].ObjectFamilyID != "evidence.blobs" {
			t.Fatalf("unexpected Evidence recovery object families: %#v", contribution.ObjectFamilies)
		}
	})

	t.Run("revisions", func(t *testing.T) {
		t.Parallel()
		contribution := RevisionProviderContribution()
		if contribution.SourceOwnerModule != revisions.SourceOwnerEvidence ||
			len(contribution.Records) != 1 {
			t.Fatalf("unexpected Evidence revision contribution: %#v", contribution)
		}
		record := contribution.Records[0]
		if record.SourceOwnerModule != revisions.SourceOwnerEvidence ||
			record.RecordType != "evidence" ||
			record.DeleteRestoreSource == nil ||
			record.RowRollbackProvider == nil {
			t.Fatalf("unexpected Evidence revision record contribution: %#v", record)
		}
		if len(record.RecordViewRoutes) != 1 ||
			record.RecordViewRoutes[0].ContributionID != "evidence.evidence" ||
			!slices.Equal(record.RecordViewRoutes[0].ViewSchemaIDs, []string{ViewSchemaID}) {
			t.Fatalf("unexpected Evidence revision routes: %#v", record.RecordViewRoutes)
		}
	})

	t.Run("record subtype presence", func(t *testing.T) {
		t.Parallel()
		contribution := IncidentBundleSubtypeContribution()
		if contribution.FamilyID != "evidence" || contribution.Source == nil ||
			!slices.Equal(contribution.Source.SupportedRecordTypes(), []subtypepresence.RecordType{subtypepresence.RecordTypeEvidence}) {
			t.Fatalf("unexpected Evidence subtype-presence contribution: %#v", contribution)
		}
	})
}

func TestEvidenceServiceConstructionRejectsIncompleteDependencies(t *testing.T) {
	t.Parallel()

	valid := blobLifecycleDependencies{
		Postgres:       constructionDB{},
		Revisions:      &revisions.Appender{},
		Projections:    constructionProjectionRows{},
		SupportEffects: constructionSupportEffects{},
		Collaboration:  collaborationsupport.NewRecordChangedAppender(),
	}
	tests := []struct {
		name   string
		mutate func(*blobLifecycleDependencies)
	}{
		{name: "postgres", mutate: func(dependencies *blobLifecycleDependencies) { dependencies.Postgres = nil }},
		{name: "revisions", mutate: func(dependencies *blobLifecycleDependencies) { dependencies.Revisions = nil }},
		{name: "projections", mutate: func(dependencies *blobLifecycleDependencies) { dependencies.Projections = nil }},
		{name: "support effects", mutate: func(dependencies *blobLifecycleDependencies) { dependencies.SupportEffects = nil }},
		{name: "collaboration", mutate: func(dependencies *blobLifecycleDependencies) { dependencies.Collaboration = nil }},
	}
	for _, test := range tests {
		t.Run("blob lifecycle requires "+test.name, func(t *testing.T) {
			dependencies := valid
			test.mutate(&dependencies)
			if _, err := newBlobLifecycleService(dependencies); err == nil {
				t.Fatalf("newBlobLifecycleService() accepted missing %s dependency", test.name)
			}
		})
	}

	if _, err := newAccessHandleService(nil); err == nil {
		t.Fatal("newAccessHandleService() accepted a missing Postgres dependency")
	}
	if _, err := newCleanupService(nil); err == nil {
		t.Fatal("newCleanupService() accepted a missing Postgres dependency")
	}
	blobs, err := newBlobLifecycleService(valid)
	if err != nil {
		t.Fatalf("newBlobLifecycleService(valid): %v", err)
	}
	access, err := newAccessHandleService(constructionDB{})
	if err != nil {
		t.Fatalf("newAccessHandleService(valid): %v", err)
	}
	if _, err := newRouteOperations(nil, access); err == nil {
		t.Fatal("newRouteOperations() accepted a missing blob lifecycle capability")
	}
	if _, err := newRouteOperations(blobs, nil); err == nil {
		t.Fatal("newRouteOperations() accepted a missing access-handle capability")
	}
	if _, err := newSourceMutationService(constructionDB{}, nil, valid.Revisions, valid.Collaboration); err == nil {
		t.Fatal("newSourceMutationService() accepted a missing Projections dependency")
	}
	if _, err := newSourceMutationService(constructionDB{}, valid.Projections, nil, valid.Collaboration); err == nil {
		t.Fatal("newSourceMutationService() accepted a missing Revisions dependency")
	}
	if _, err := newSourceMutationService(nil, valid.Projections, valid.Revisions, valid.Collaboration); err == nil {
		t.Fatal("newSourceMutationService() accepted a missing Postgres dependency")
	}
	if _, err := newSourceMutationService(constructionDB{}, valid.Projections, valid.Revisions, nil); err == nil {
		t.Fatal("newSourceMutationService() accepted a missing Collaboration dependency")
	}
}

func TestEvidenceOwnerRuntimeRejectsIncompleteDependencies(t *testing.T) {
	t.Parallel()

	codec := conflicttokens.ConflictTokenCodec{}
	valid := OwnerRuntimeDependencies{
		Postgres:            constructionDB{},
		ConflictTokens:      &codec,
		Revisions:           &revisions.Appender{},
		Collaboration:       collaborationsupport.NewRecordChangedAppender(),
		ObjectStore:         constructionObjectStore{},
		ConflictFields:      constructionConflictFields{},
		ConflictIdempotency: constructionConflictIdempotency{},
		CleanupObserver:     &dispatcherTestObserver{},
		Now:                 time.Now,
		Projections: evidenceprojection.Ports{
			Rows:           constructionProjectionRows{},
			SupportEffects: constructionSupportEffects{},
		},
	}
	tests := []struct {
		name   string
		mutate func(*OwnerRuntimeDependencies)
	}{
		{name: "postgres", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Postgres = nil }},
		{name: "conflict tokens", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.ConflictTokens = nil }},
		{name: "revisions", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Revisions = nil }},
		{name: "collaboration", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Collaboration = nil }},
		{name: "object store", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.ObjectStore = nil }},
		{name: "conflict fields", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.ConflictFields = nil }},
		{name: "conflict idempotency", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.ConflictIdempotency = nil }},
		{name: "projection rows", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Projections.Rows = nil }},
		{name: "support effects", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Projections.SupportEffects = nil }},
		{name: "cleanup observer", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.CleanupObserver = nil }},
		{name: "clock", mutate: func(dependencies *OwnerRuntimeDependencies) { dependencies.Now = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := valid
			test.mutate(&dependencies)
			if _, err := NewOwnerRuntime(dependencies); err == nil {
				t.Fatalf("NewOwnerRuntime() accepted missing %s dependency", test.name)
			}
		})
	}

	runtime, err := NewOwnerRuntime(valid)
	if err != nil {
		t.Fatalf("NewOwnerRuntime(valid): %v", err)
	}
	if runtime.RouteRegistrar(Settings{}) == nil || runtime.MutationContribution() == nil ||
		runtime.TimelineAttachmentContribution() == nil || runtime.ImportCreateFacade() == nil {
		t.Fatal("NewOwnerRuntime(valid) returned an incomplete capability set")
	}
	if runtime.CleanupDispatcher() == nil {
		t.Fatal("NewOwnerRuntime(valid) omitted the cleanup dispatcher")
	}
	binding := runtime.ImportCreateFacade().ImportOwnerCreateBinding()
	if binding.TargetViewSchemaID != ViewSchemaID || binding.FacadeID != "evidence.import_create" {
		t.Fatalf("Evidence import binding = %#v", binding)
	}
}

type constructionDB struct{}

func (constructionDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (constructionDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (constructionDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (constructionDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, nil
}

type constructionProjectionRows struct{}

func (constructionProjectionRows) RefreshEvidenceTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func (constructionProjectionRows) LoadEvidenceTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	return nil, nil
}

type constructionSupportEffects struct{}

func (constructionSupportEffects) RefreshEvidenceAssociationEffects(
	context.Context,
	pgx.Tx,
	evidenceprojection.EvidenceAssociationEffectsInput,
) (evidenceprojection.EvidenceAssociationEffectsResult, error) {
	return evidenceprojection.EvidenceAssociationEffectsResult{}, nil
}

type constructionObjectStore struct {
	objectstore.Store
	objectstore.TypedStore
}

type constructionConflictFields struct {
	conflicttokens.FieldResolver
}

type constructionConflictIdempotency struct {
	conflicttokens.IdempotencyPort
}
