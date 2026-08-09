package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestRestoreProjectionRebuildNoProvidersIsNotApplicable(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "projection-restore-no-providers")
	rebuilder := NewRestoreRebuilderFromStore(NewStore(db, nil))

	result, err := rebuilder.RebuildRestoreProjections(t.Context(), validCharacterizationRebuildRequest())
	if err != nil {
		t.Fatalf("no-provider restore projection rebuild: %v", err)
	}
	if result.Status != restorecontract.ProjectionRebuildStatusNotApplicable ||
		result.ReadinessOutcome != restorecontract.ProjectionReadinessNotApplicable ||
		!result.ReadinessSatisfied() || len(result.ProviderResults) != 0 {
		t.Fatalf("no-provider result = %#v", result)
	}
}

func TestRestoreProjectionRebuildParticipationMatrix(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "projection-restore-participation")

	t.Run("nonparticipating is explicit and not applicable", func(t *testing.T) {
		catalog := mustCharacterizationCatalog(t, []Provider{{
			descriptor: characterizationProviderDescriptor(
				"nonparticipating",
				"cartulary.view.characterization.nonparticipating.v1",
				"host_grid_projection",
				nil,
				ProviderStatusActive,
				RestoreRebuildNonparticipating,
			),
		}})
		result, err := NewRestoreRebuilder(db, catalog).RebuildRestoreProjections(
			t.Context(),
			validCharacterizationRebuildRequest(),
		)
		if err != nil {
			t.Fatalf("nonparticipating restore projection rebuild: %v", err)
		}
		if result.Status != restorecontract.ProjectionRebuildStatusNotApplicable ||
			result.ReadinessOutcome != restorecontract.ProjectionReadinessNotApplicable ||
			len(result.ProviderResults) != 1 ||
			result.ProviderResults[0].Status != restorecontract.ProjectionProviderResultSkippedNonparticipating {
			t.Fatalf("nonparticipating result = %#v", result)
		}
	})

	t.Run("unsupported provider fails closed", func(t *testing.T) {
		catalog := mustCharacterizationCatalog(t, []Provider{{
			descriptor: characterizationProviderDescriptor(
				"unsupported",
				"cartulary.view.characterization.unsupported.v1",
				"host_grid_projection",
				nil,
				ProviderStatusDeprecated,
				RestoreRebuildUnsupported,
			),
		}})
		result, err := NewRestoreRebuilder(db, catalog).RebuildRestoreProjections(
			t.Context(),
			validCharacterizationRebuildRequest(),
		)
		if err == nil || !strings.Contains(err.Error(), "not restore-rebuild ready") {
			t.Fatalf("unsupported provider error = %v", err)
		}
		if result.Status != restorecontract.ProjectionRebuildStatusFailed ||
			result.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete ||
			len(result.ProviderResults) != 1 ||
			result.ProviderResults[0].Status != restorecontract.ProjectionProviderResultFailed {
			t.Fatalf("unsupported provider result = %#v", result)
		}
	})
}

func TestRestoreProjectionRebuildProviderFailureRollsBackEveryPosition(t *testing.T) {
	for _, failAt := range []int{0, 1, 2} {
		failAt := failAt
		t.Run(fmt.Sprintf("position_%d", failAt+1), func(t *testing.T) {
			db := pgtest.Start(t).BeginRollbackDBT(t, fmt.Sprintf("projection-restore-provider-failure-%d", failAt))
			actorID := uuid.New()
			incidentID := uuid.New()
			recordID := uuid.New()
			if _, err := db.Exec(t.Context(), `
INSERT INTO users (
    id, email, display_name, password_hash, mfa_required, is_active,
    is_deployment_admin, created_at, updated_at
) VALUES ($1, $2, 'Projection Failure', 'hash', false, true, false, now(), now())
`, actorID, fmt.Sprintf("projection-failure-%d@example.test", failAt)); err != nil {
				t.Fatalf("seed restore failure actor: %v", err)
			}
			if _, err := db.Exec(t.Context(), `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id, created_at, updated_at
) VALUES ($1, $2, $2, 'Projection restore failure atomicity', 'active', $3, $3, now(), now())
`, incidentID, fmt.Sprintf("IR-PROJECTION-FAILURE-%d", failAt), actorID); err != nil {
				t.Fatalf("seed restore failure incident: %v", err)
			}
			if _, err := db.Exec(t.Context(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_by_user_id, updated_by_user_id,
    created_at, updated_at, row_version
) VALUES ($1, $2, 'host', $3, $3, now(), now(), 1)
`, recordID, incidentID, actorID); err != nil {
				t.Fatalf("seed restore failure record: %v", err)
			}
			if _, err := db.Exec(t.Context(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    created_by_user_id, updated_by_user_id, created_at, updated_at, row_version
) VALUES ($1, $2, 'before', 'before', 'canonical', $3, $3, now(), now(), 1)
`, recordID, incidentID, actorID); err != nil {
				t.Fatalf("seed restore failure host: %v", err)
			}
			if _, err := db.Exec(t.Context(), `
INSERT INTO host_grid_projection (
    record_id, incident_id, row_version, display_name, hostname, host_state, edited_at
) VALUES ($1, $2, 1, 'before', 'before', 'canonical', now())
`, recordID, incidentID); err != nil {
				t.Fatalf("seed restore failure projection: %v", err)
			}

			calls := make([]string, 0, 3)
			providers := make([]Provider, 0, 3)
			for index, key := range []string{"first", "middle", "final"} {
				index := index
				key := key
				after := []string(nil)
				if index > 0 {
					after = []string{[]string{"first", "middle"}[index-1]}
				}
				descriptor := characterizationProviderDescriptor(
					key,
					"cartulary.view.characterization."+key+".v1",
					"host_grid_projection",
					after,
					ProviderStatusActive,
					RestoreRebuildRequired,
				)
				descriptor.Capabilities = ProviderCapabilities{IncidentRebuild: true, RestoreRebuild: true}
				providers = append(providers, Provider{
					descriptor: descriptor,
					rebuildIncidentTx: func(ctx context.Context, _ *Store, tx pgx.Tx, gotIncidentID uuid.UUID) error {
						calls = append(calls, key)
						if gotIncidentID != incidentID {
							return fmt.Errorf("provider %s incident = %s", key, gotIncidentID)
						}
						if _, err := tx.Exec(ctx, `UPDATE host_grid_projection SET display_name = $1 WHERE record_id = $2`, key, recordID); err != nil {
							return err
						}
						if index == failAt {
							return errors.New("injected provider failure")
						}
						return nil
					},
				})
			}

			result, err := NewRestoreRebuilder(db, mustCharacterizationCatalog(t, providers)).RebuildRestoreProjections(
				t.Context(),
				validCharacterizationRebuildRequest(),
			)
			if err == nil || !strings.Contains(err.Error(), "injected provider failure") {
				t.Fatalf("provider failure error = %v", err)
			}
			wantCalls := []string{"first", "middle", "final"}[:failAt+1]
			if !reflect.DeepEqual(calls, wantCalls) {
				t.Fatalf("provider calls got %#v want %#v", calls, wantCalls)
			}
			if len(result.ProviderResults) != failAt+1 {
				t.Fatalf("provider result prefix = %#v", result.ProviderResults)
			}
			for _, providerResult := range result.ProviderResults {
				if providerResult.Status != restorecontract.ProjectionProviderResultFailed ||
					len(providerResult.RebuiltViewSchemaIDs) != 0 ||
					len(providerResult.RebuiltProjectionTables) != 0 {
					t.Fatalf("failed restore claimed rebuilt resources: %#v", providerResult)
				}
			}
			var displayName string
			if err := db.QueryRow(t.Context(), `SELECT display_name FROM host_grid_projection WHERE record_id = $1`, recordID).Scan(&displayName); err != nil {
				t.Fatalf("load rolled-back projection: %v", err)
			}
			if displayName != "before" {
				t.Fatalf("provider failure retained projection mutation %q", displayName)
			}
		})
	}
}

func TestRestoreProjectionRebuildCancellationAndSourceReferenceValidation(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "projection-restore-cancellation")
	descriptor := characterizationProviderDescriptor(
		"required",
		"cartulary.view.characterization.required.v1",
		"host_grid_projection",
		nil,
		ProviderStatusActive,
		RestoreRebuildRequired,
	)
	descriptor.Capabilities = ProviderCapabilities{IncidentRebuild: true, RestoreRebuild: true}
	provider := Provider{
		descriptor: descriptor,
		rebuildIncidentTx: func(context.Context, *Store, pgx.Tx, uuid.UUID) error {
			t.Fatal("canceled or invalid restore reached provider")
			return nil
		},
	}
	rebuilder := NewRestoreRebuilder(db, mustCharacterizationCatalog(t, []Provider{provider}))

	invalidRef := validCharacterizationRebuildRequest()
	invalidRef.RestoredSourceStateRef = "   "
	result, err := rebuilder.RebuildRestoreProjections(t.Context(), invalidRef)
	if err == nil || !strings.Contains(err.Error(), "restored_source_state_ref is required") ||
		result.Errors[0].Code != "invalid_restore_projection_rebuild_request" {
		t.Fatalf("invalid source-state reference result=%#v error=%v", result, err)
	}

	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = rebuilder.RebuildRestoreProjections(canceled, validCharacterizationRebuildRequest())
	if err == nil || !errors.Is(err, context.Canceled) ||
		result.Status != restorecontract.ProjectionRebuildStatusFailed ||
		result.ReadinessOutcome != restorecontract.ProjectionReadinessIncomplete ||
		len(result.ProviderResults) != 0 {
		t.Fatalf("canceled restore result=%#v error=%v", result, err)
	}
}

func characterizationProviderDescriptor(
	providerKey string,
	viewSchemaID string,
	table string,
	rebuildAfter []string,
	status ProviderStatus,
	participation RestoreRebuildParticipation,
) ProviderDescriptor {
	return ProviderDescriptor{
		SchemaVersion:             providercontract.DescriptorSchemaVersion,
		Status:                    status,
		ProviderKey:               providerKey,
		SourceOwnerKey:            "entities",
		ViewSchemaIDs:             []string{viewSchemaID},
		SourceRecordTypes:         []string{providerKey},
		SourceAuthorityModules:    []string{"entities"},
		ProjectionTableFamilies:   []string{table},
		ProjectionStorageOwnerKey: "projections",
		RestoreRebuild:            participation,
		FacadePackages:            []string{"internal/modules/entities"},
		RebuildAfter:              rebuildAfter,
	}
}

func mustCharacterizationCatalog(t testing.TB, providers []Provider) *Catalog {
	t.Helper()
	catalog, err := NewCatalog(providers)
	if err != nil {
		t.Fatalf("characterization catalog: %v", err)
	}
	return catalog
}

func validCharacterizationRebuildRequest() restorecontract.ProjectionRebuildRequest {
	return restorecontract.ProjectionRebuildRequest{
		RestoreOperationID:     uuid.New(),
		RestoredSourceStateRef: "backup_set:" + uuid.NewString(),
		RebuildScope:           restorecontract.ProjectionRebuildScopeAllActiveProviders,
		ProviderRegistryRef:    restorecontract.ProviderRegistryRefCodeBacked,
	}
}
