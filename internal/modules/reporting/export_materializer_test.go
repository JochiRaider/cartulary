package reporting

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

// Exercise the provider port rather than private adapter names or source files.
// The transaction deliberately implements no database operations: any SQL or
// transaction control by the coordinator fails this test, while the source
// owners receive the exact borrowed transaction and context.
func TestReportingMaterializerUsesOwnerContracts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tx := &borrowedReportingTx{}
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000700002")
	refs := map[string][]string{"source": {"/evidence/support"}}
	var calls []string
	checkInputs := func(gotCtx context.Context, gotTx pgx.Tx, gotIncident uuid.UUID) {
		t.Helper()
		if gotCtx != ctx || gotTx != tx || gotIncident != incidentID {
			t.Fatal("provider did not receive the borrowed snapshot context")
		}
	}
	output := func(key, path, id string) exportprovider.ProviderOutput {
		t.Helper()
		value, err := exportprovider.NewProviderOutput(key, []exportprovider.BoundField{{
			RecordID: id, RecordType: "evidence",
			FieldFact: exportprovider.FieldFact{
				SchemaID: exportprovider.FieldFactSchemaID, Path: path,
				SourceFamily: "evidence", ContentClass: ContentClassSourceEvidence,
				Value: "source metadata", DisclosurePartitionRefs: []string{"shared"},
				SupportRefs: []string{"/evidence/support"},
			},
		}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	first := output("first", "/z/opaque-selector", "declared-first")
	second := output("second", "/a/another-selector", "declared-second")
	provider := func(value exportprovider.ProviderOutput) exportprovider.FieldProvider {
		return reportingExportFieldProviderFunc{key: value.ProviderKey, collect: func(gotCtx context.Context, gotTx pgx.Tx, gotIncident uuid.UUID, gotRefs map[string][]string) (exportprovider.ProviderOutput, error) {
			checkInputs(gotCtx, gotTx, gotIncident)
			if !reflect.DeepEqual(gotRefs, refs) {
				t.Fatal("provider lost support-owner contribution")
			}
			calls = append(calls, value.ProviderKey)
			return value, nil
		}}
	}
	materializer := reportingExportMaterializer{
		supportRefProvider: supportReferenceProbe{collect: func(gotCtx context.Context, gotTx pgx.Tx, gotIncident uuid.UUID) (map[string][]string, error) {
			checkInputs(gotCtx, gotTx, gotIncident)
			calls = append(calls, "support")
			return refs, nil
		}},
		fieldProviders: []exportprovider.FieldProvider{provider(first), provider(second)},
	}
	source, err := materializer.CollectSourceTx(ctx, tx, incidentID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []string{"support", "first", "second"}) {
		t.Fatalf("owner calls = %v", calls)
	}
	if len(source.Fields) != 2 || source.Fields[0].Path != second.FieldFacts[0].Path || source.Fields[1].Path != first.FieldFacts[0].Path || source.Fields[0].SourceRecordID != "declared-second" || source.Fields[1].SourceRecordID != "declared-first" {
		t.Fatalf("source identities or deterministic selection lost: %#v", source)
	}
	if source.Fields[0].SourceFamily != "evidence" || source.Fields[0].SourceRecordType != "evidence" || source.Fields[0].ContentClass != ContentClassSourceEvidence || source.Fields[0].Value != "source metadata" {
		t.Fatalf("owner facts changed: %#v", source.Fields[0])
	}
	source.Fields[0].DisclosurePartitionRefs[0] = "changed"
	source.Fields[0].SupportRefs[0] = "/changed"
	if second.FieldFacts[0].DisclosurePartitionRefs[0] != "shared" || second.FieldFacts[0].SupportRefs[0] != "/evidence/support" {
		t.Fatal("materialization aliased owner metadata")
	}

	t.Run("support failure precedes all field reads", func(t *testing.T) {
		failure := errors.New("support owner failed")
		copy := materializer
		copy.supportRefProvider = supportReferenceProbe{collect: func(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error) { return nil, failure }}
		calls = nil
		source, err := copy.CollectSourceTx(ctx, tx, incidentID)
		if !errors.Is(err, failure) || len(source.Fields) != 0 || len(source.Timeline) != 0 || len(calls) != 0 {
			t.Fatalf("support failure did not stop materialization: fields=%v calls=%v err=%v", source, calls, err)
		}
	})
	for _, name := range []string{"owner failure", "wrong owner", "invalid binding", "duplicate selector"} {
		t.Run(name, func(t *testing.T) {
			failure := errors.New("source owner failed")
			bad := output("invalid", "/bad/selector", "bad-record")
			var sourceErr error
			switch name {
			case "owner failure":
				sourceErr = failure
			case "wrong owner":
				bad.ProviderKey = "another-owner"
			case "invalid binding":
				bad.RecordFacts[0].FieldPaths = []string{"/missing"}
			case "duplicate selector":
				bad = output("invalid", first.FieldFacts[0].Path, "different-record")
			}
			copy := materializer
			copy.fieldProviders = []exportprovider.FieldProvider{provider(first), reportingExportFieldProviderFunc{
				key: "invalid", collect: func(context.Context, pgx.Tx, uuid.UUID, map[string][]string) (exportprovider.ProviderOutput, error) {
					return bad, sourceErr
				},
			}, reportingExportFieldProviderFunc{
				key: "unreached", collect: func(context.Context, pgx.Tx, uuid.UUID, map[string][]string) (exportprovider.ProviderOutput, error) {
					t.Fatal("provider called after failed contribution")
					return exportprovider.ProviderOutput{}, nil
				},
			}}
			source, err := copy.CollectSourceTx(ctx, tx, incidentID)
			if err == nil || len(source.Fields) != 0 || len(source.Timeline) != 0 || !strings.Contains(err.Error(), "provider invalid") || (sourceErr != nil && !errors.Is(err, failure)) {
				t.Fatalf("invalid contribution published partial output or lost owner failure: fields=%v err=%v", source, err)
			}
		})
	}
}

func TestReportingMaterializerRejectsInvalidOwnerRegistration(t *testing.T) {
	boundary := &borrowedSourceBoundary{}
	support := staticSupportRefProvider{}
	if _, err := newReportingExportMaterializer(boundary, support, reportingExportFieldProviderFunc{}); err == nil {
		t.Fatal("empty provider key accepted")
	}
	base, err := newReportingExportMaterializer(boundary, support)
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range base.fieldProviders {
		if _, err := newReportingExportMaterializer(boundary, support, provider); err == nil {
			t.Fatalf("duplicate owner %q accepted", provider.ProviderKey())
		}
	}
	custom := reportingExportFieldProviderFunc{key: "custom"}
	if _, err := newReportingExportMaterializer(boundary, support, custom, custom); err == nil {
		t.Fatal("duplicate contributed owner accepted")
	}
	if _, err := newReportingExportMaterializer(nil, support); err == nil {
		t.Fatal("missing source boundary accepted")
	}
	if _, err := newReportingExportMaterializer(boundary, nil); err == nil {
		t.Fatal("missing support owner accepted")
	}
	if _, err := newReportingExportMaterializer(boundary, support, nil); err == nil {
		t.Fatal("missing field owner accepted")
	}
}

type borrowedReportingTx struct{ pgx.Tx }
type borrowedSourceBoundary struct{ SourceBoundaryResolver }

type supportReferenceProbe struct {
	collect func(context.Context, pgx.Tx, uuid.UUID) (map[string][]string, error)
}

func (supportReferenceProbe) ProviderKey() string { return "support" }
func (p supportReferenceProbe) CollectSupportRefsTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (map[string][]string, error) {
	return p.collect(ctx, tx, id)
}
