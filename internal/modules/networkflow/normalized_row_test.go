package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertCanonicalRowConstruction(t *testing.T) {
	t.Helper()
	params := createTableParams{
		IncidentID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		ActorUserID:         uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		ImportSessionID:     uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		ImportUnitID:        uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		SourceContentSHA256: strings.Repeat("2", 64), MappingFingerprint: strings.Repeat("3", 64),
		SourceFilenameDigest: strings.Repeat("5", 64), SourceFilenameDigestKeyID: "constructor-test",
		SourceProfileID: sourceProfileCiscoSNANetFlowCSV, ParserProfileID: parserProfileRFC4180HeaderedCSV,
		RowCountAccepted: 1, Now: time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC),
		VisitRows: func(func(normalizedRow) error) error { t.Fatal("invalid context invoked row visitor"); return nil },
	}
	t.Run("invalid context rejected before transaction access", func(t *testing.T) {
		cases := map[string]func(*createTableParams){
			"incident":        func(p *createTableParams) { p.IncidentID = uuid.Nil },
			"actor":           func(p *createTableParams) { p.ActorUserID = uuid.Nil },
			"session":         func(p *createTableParams) { p.ImportSessionID = uuid.Nil },
			"unit":            func(p *createTableParams) { p.ImportUnitID = uuid.Nil },
			"source digest":   func(p *createTableParams) { p.SourceContentSHA256 = "invalid" },
			"mapping digest":  func(p *createTableParams) { p.MappingFingerprint = strings.Repeat("A", 64) },
			"filename digest": func(p *createTableParams) { p.SourceFilenameDigest = "" },
			"digest key":      func(p *createTableParams) { p.SourceFilenameDigestKeyID = "invalid key" },
			"source profile":  func(p *createTableParams) { p.SourceProfileID = "unknown" },
			"missing parser":  func(p *createTableParams) { p.ParserProfileID = "" },
			"wrong parser":    func(p *createTableParams) { p.ParserProfileID = "unknown" },
			"missing time":    func(p *createTableParams) { p.Now = time.Time{} },
			"invalid time":    func(p *createTableParams) { p.Now = time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC) },
		}
		for name, mutate := range cases {
			t.Run(name, func(t *testing.T) {
				invalid := params
				mutate(&invalid)
				// A nil transaction makes any accidental database access fail immediately.
				if _, err := (&store{}).CreateTableTx(context.Background(), nil, invalid); !errors.Is(err, errInvalidImportContext) {
					t.Fatalf("invalid import context: %v", err)
				}
			})
		}
	})
	rowContext, err := validatedRowImportContext(params)
	if err != nil {
		t.Fatal(err)
	}
	const tableID = "nft_0123456789abcdef0123456789abcdef"
	source := normalizedRow{
		SourceRowNumber: 2, SourceRowDigestSHA256: strings.Repeat("1", 64), NormalizedRowDigestSHA256: strings.Repeat("4", 64),
		FlowStartUTC: time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), FlowEndUTC: time.Date(1, 1, 1, 0, 0, 1, 0, time.UTC),
		SrcIP: "192.0.2.1", DstIP: "2001:db8::1", IPProtocol: 6, BytesCount: "0", PacketsCount: "0", UnmappedRaw: json.RawMessage(`{}`),
	}
	t.Run("identity provenance and nullable values", func(t *testing.T) {
		row, err := rowContext.persistedRow(tableID, source)
		if err != nil {
			t.Fatal(err)
		}
		// Independent fixed NF-040 vectors, including identical semantic values at
		// distinct source locators. These are not outputs captured from the factory.
		const expectedID = "nfr_bd499b00a9e5db32b3145b98c33084f7882c5356d428ed95dc0e011295887e1f"
		if row.RowID != expectedID || row.IncidentID != params.IncidentID || row.NetworkFlowTableID != tableID ||
			row.MappingFingerprint != params.MappingFingerprint || row.CreatedByUserID != params.ActorUserID || !row.CreatedAt.Equal(params.Now) {
			t.Fatalf("storage-owned identity or attribution changed: %#v", row)
		}
		expectedSource := source
		actualSource := normalizedRow{
			SourceRowNumber: row.SourceRowNumber, SourceRowDigestSHA256: row.SourceRowDigestSHA256, NormalizedRowDigestSHA256: row.NormalizedRowDigestSHA256,
			FlowStartUTC: row.FlowStartUTC, FlowEndUTC: row.FlowEndUTC, SrcIP: row.SrcIP, DstIP: row.DstIP, SrcPort: row.SrcPort, DstPort: row.DstPort,
			IPProtocol: row.IPProtocol, BytesCount: row.BytesCount, PacketsCount: row.PacketsCount, ExporterID: row.ExporterID,
			InputInterface: row.InputInterface, OutputInterface: row.OutputInterface, TCPFlags: row.TCPFlags, ApplicationLabel: row.ApplicationLabel, UnmappedRaw: row.UnmappedRaw,
		}
		if !reflect.DeepEqual(actualSource, expectedSource) {
			t.Fatalf("normalized values changed: %#v", actualSource)
		}
		var provenance map[string]any
		if err := json.Unmarshal(row.ObservationSourceRef, &provenance); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{
			"import_session_id": params.ImportSessionID.String(), "import_unit_id": params.ImportUnitID.String(),
			"source_content_sha256": params.SourceContentSHA256, "source_profile_id": params.SourceProfileID, "parser_profile_id": params.ParserProfileID,
			"mapping_fingerprint": params.MappingFingerprint, "source_row_number": float64(2), "source_row_digest_sha256": source.SourceRowDigestSHA256,
		}
		if !reflect.DeepEqual(provenance, want) {
			t.Fatalf("provenance: %#v", provenance)
		}
		duplicate := source
		duplicate.SourceRowNumber = 3
		next, err := rowContext.persistedRow(tableID, duplicate)
		if err != nil || next.RowID != "nfr_060e6451ce29704f3da16852b1c3b9492f09259ff2860038472b40a0dbbeb594" {
			t.Fatalf("distinct source locator lost: %#v %v", next, err)
		}
	})
	t.Run("malformed construction fails", func(t *testing.T) {
		for _, id := range []string{"", "nft_bad", "nft_" + strings.Repeat("G", 32), "nft_" + strings.Repeat("A", 32)} {
			if _, err := rowContext.persistedRow(id, source); !errors.Is(err, errInvalidImportContext) {
				t.Fatalf("invalid table identity %q: %v", id, err)
			}
		}
		cases := map[string]func(*normalizedRow){
			"locator":           func(r *normalizedRow) { r.SourceRowNumber = 0 },
			"source digest":     func(r *normalizedRow) { r.SourceRowDigestSHA256 = "" },
			"normalized digest": func(r *normalizedRow) { r.NormalizedRowDigestSHA256 = "invalid" },
			"missing raw":       func(r *normalizedRow) { r.UnmappedRaw = nil },
			"null raw":          func(r *normalizedRow) { r.UnmappedRaw = json.RawMessage(`null`) },
			"malformed raw":     func(r *normalizedRow) { r.UnmappedRaw = json.RawMessage(`{"1":null}`) },
			"duplicate raw":     func(r *normalizedRow) { r.UnmappedRaw = json.RawMessage(`{"1":null,"1":null}`) },
		}
		for name, mutate := range cases {
			t.Run(name, func(t *testing.T) {
				invalid := source
				mutate(&invalid)
				if _, err := rowContext.persistedRow(tableID, invalid); !errors.Is(err, errInvalidNormalizedRow) {
					t.Fatalf("invalid normalized input: %v", err)
				}
			})
		}
	})
}
