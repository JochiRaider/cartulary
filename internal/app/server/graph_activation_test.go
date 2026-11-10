package server

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type graphActivationNoIO struct{ dbport.DB }

func graphActivationControl() recovery.TargetControl {
	text := func(s string) *string { return &s }
	hash := strings.Repeat("a", 64)
	return recovery.TargetControl{
		SchemaID: recovery.TargetStateSchemaID, State: recovery.TargetComplete, Revision: 7,
		TargetGenerationID: text(uuid.NewString()), Purpose: text("restore_target"),
		Operation: text("restore_latest"), OperationID: text(uuid.NewString()), BackupSetID: text(uuid.NewString()),
		RecoveryStateCatalogSHA256: text(graphrestore.CurrentRestoreImplementationBinding().Binding.RecoveryStateCatalogSHA256),
		CodecRegistrySHA256:        &hash, CompletionEvidenceSHA256: &hash,
		BindingDigests: &recovery.TargetBindingDigests{DatabaseSHA256: hash, ObjectStoreSHA256: hash, ReferencePackStorageSHA256: hash, ExportOutputsSHA256: hash},
	}
}

func graphActivationSuccess(request graphrestore.RestoreRebuildRequest) graphrestore.RestoreRebuildResult {
	hash := strings.Repeat("b", 64)
	return graphrestore.RestoreRebuildResult{
		SchemaID: graphrestore.RestoreRebuildResultSchemaID, AlgorithmID: graphrestore.RestoreAlgorithmID,
		RestoreOperationID: request.RestoreOperationID.String(), TargetGenerationID: request.TargetGenerationID.String(),
		Status: graphrestore.RestoreStatusSucceeded, ReadinessOutcome: graphrestore.RestoreReadinessReady,
		ImplementationBindingSHA256: request.ImplementationBinding.SHA256, SourceRegistrySHA256: request.SourceRegistry.SHA256,
		ClearedTableIDs: graphrestore.RestoreGraphTableIDs(), RebuiltViews: []graphrestore.RestoreRebuiltView{},
		PostconditionSHA256: &hash, Warnings: []graphrestore.RestoreSafeMessage{}, Errors: []graphrestore.RestoreSafeMessage{},
	}
}

func awaitGraphActivation(t *testing.T, activation *claimedGraphActivation) {
	t.Helper()
	select {
	case <-activation.done:
	case <-time.After(5 * time.Second):
		t.Fatal("Graph activation did not finish")
	}
}

func TestClaimedGraphActivation_Unit(t *testing.T) {
	for _, name := range []string{"success", "failure", "unknown commit", "panic", "wrong generation", "wrong operation", "wrong binding", "invalid result", "cancelled"} {
		t.Run(name, func(t *testing.T) {
			control := graphActivationControl()
			var calls, fatals atomic.Int32
			activation, err := newClaimedGraphActivation(graphActivationNoIO{}, control, true, time.Now, func() { fatals.Add(1) })
			if err != nil {
				t.Fatal(err)
			}
			if activation.Ready() || activation.expected.RestoreOperationID.String() != *control.OperationID || activation.expected.TargetGenerationID.String() != *control.TargetGenerationID {
				t.Fatal("restored activation lost original identity or opened before reconstruction")
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			activation.rebuild = func(context.Context) (graphrestore.RestoreRebuildResult, error) {
				calls.Add(1)
				result := graphActivationSuccess(*activation.expected)
				switch name {
				case "failure":
					return result, graphrestore.NewRestoreError(graphrestore.RestoreErrorPublicationFailed)
				case "unknown commit":
					return result, graphrestore.NewRestoreError(graphrestore.RestoreErrorOutcomeIndeterminate)
				case "panic":
					panic("SECRET callback panic")
				case "wrong generation":
					result.TargetGenerationID = uuid.NewString()
				case "wrong operation":
					result.RestoreOperationID = uuid.NewString()
				case "wrong binding":
					result.ImplementationBindingSHA256 = strings.Repeat("c", 64)
				case "invalid result":
					result.ClearedTableIDs = nil
				case "cancelled":
					cancel()
				}
				return result, nil
			}
			serving := make(chan struct{})
			if err := activation.Prepare(ctx, serving); err != nil {
				t.Fatal(err)
			}
			if err := activation.Prepare(ctx, serving); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 0 {
				t.Fatal("source callback ran before publication")
			}
			close(serving)
			awaitGraphActivation(t, activation)
			if err := activation.Prepare(ctx, serving); err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 || activation.Ready() != (name == "success") {
				t.Fatalf("calls=%d ready=%v", calls.Load(), activation.Ready())
			}
			wantFatal := int32(0)
			if name == "unknown commit" || name == "panic" {
				wantFatal = 1
			}
			if fatals.Load() != wantFatal {
				t.Fatalf("fatal count=%d", fatals.Load())
			}
			select {
			case <-activation.dependentReady:
				if name != "success" {
					t.Fatal("dependent cleanup gate opened")
				}
			default:
				if name == "success" {
					t.Fatal("dependent cleanup gate remained closed")
				}
			}
			if err := activation.Close(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, claimed := range []bool{false, true} {
		t.Run(map[bool]string{false: "unclaimed", true: "initialized"}[claimed], func(t *testing.T) {
			control := graphActivationControl()
			if claimed {
				control = recovery.TargetControl{SchemaID: recovery.TargetStateSchemaID, State: recovery.TargetInitialized, Revision: 1}
			}
			activation, err := newClaimedGraphActivation(graphActivationNoIO{}, control, claimed, time.Now, func() { t.Error("unexpected fatal") })
			if err != nil {
				t.Fatal(err)
			}
			if activation.rebuild != nil || activation.Ready() != claimed {
				t.Fatal("ordinary/unclaimed activation created reconstruction")
			}
			select {
			case <-activation.dependentReady:
				t.Fatal("cleanup admitted before serving")
			default:
			}
			serving := make(chan struct{})
			close(serving)
			if err := activation.Prepare(t.Context(), serving); err != nil {
				t.Fatal(err)
			}
			awaitGraphActivation(t, activation)
			if err := activation.Close(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
	}
	control := graphActivationControl()
	control.Purpose = new(string)
	*control.Purpose = "restore_verification_target"
	*control.Operation = "restore_verify_latest"
	if _, err := claimedGraphRebuildRequest(control); err == nil {
		t.Fatal("verification target admitted for application reconstruction")
	}
}
