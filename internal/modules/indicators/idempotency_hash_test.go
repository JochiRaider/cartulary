package indicators

import (
	"bytes"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
)

func testIndicatorReplayHashCompatibility(t *testing.T) {
	t.Helper()
	parsedType := "domain_name"
	targetID := uuid.MustParse("00000000-0000-4000-8000-000000000222")
	confidence := 80
	rationale := "reviewed"
	normalized := "203.0.113.7"
	defanged := "203[.]0[.]113[.]7"
	hashAlgorithm := "sha256"
	hashValue := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	stixPattern := "[file:hashes.SHA-256 = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa']"
	tests := []struct {
		name     string
		preimage []byte
		digest   []byte
		wantJSON string
		wantHash string
	}{
		{
			name: "create",
			preimage: createIndicatorRequestPreimage(CreateCommand{
				ClientTxnID: "txn-indicator", IndicatorType: "ipv4_addr",
				ValueKind: "atomic", DisplayValue: "203[.]0[.]113[.]7",
			}),
			digest: createIndicatorRequestHash(CreateCommand{
				ClientTxnID: "txn-indicator", IndicatorType: "ipv4_addr",
				ValueKind: "atomic", DisplayValue: "203[.]0[.]113[.]7",
			}),
			wantJSON: `{"client_txn_id":"txn-indicator","indicator.display_value":"203[.]0[.]113[.]7","indicator.indicator_type":"ipv4_addr","indicator.value_kind":"atomic","view_schema_id":"cartulary.view.indicators.v1"}`,
			wantHash: "49dd4b43356f985be78b671d6b57cfe912dcfc2782573acc9fb6c2cda8b5e6a6",
		},
		{
			name: "create with every optional representation",
			preimage: createIndicatorRequestPreimage(CreateCommand{
				ClientTxnID: "txn-indicator-full", IndicatorType: "file_hash",
				ValueKind: "hash", DisplayValue: "203.0.113.7", NormalizedValue: &normalized,
				DefangedValue: &defanged, HashAlgorithm: &hashAlgorithm,
				HashValue: &hashValue, STIXPattern: &stixPattern,
			}),
			digest: createIndicatorRequestHash(CreateCommand{
				ClientTxnID: "txn-indicator-full", IndicatorType: "file_hash",
				ValueKind: "hash", DisplayValue: "203.0.113.7", NormalizedValue: &normalized,
				DefangedValue: &defanged, HashAlgorithm: &hashAlgorithm,
				HashValue: &hashValue, STIXPattern: &stixPattern,
			}),
			wantJSON: `{"client_txn_id":"txn-indicator-full","indicator.defanged_value":"203[.]0[.]113[.]7","indicator.display_value":"203.0.113.7","indicator.hash_algorithm":"sha256","indicator.hash_value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","indicator.indicator_type":"file_hash","indicator.normalized_value":"203.0.113.7","indicator.stix_pattern":"[file:hashes.SHA-256 = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa']","indicator.value_kind":"hash","view_schema_id":"cartulary.view.indicators.v1"}`,
			wantHash: "df1a4a2a71c958e95e7ee7c75de14c4d52b3d95ffef1fb9c04406bde53679e75",
		},
		{
			name: "observation create",
			preimage: observationCreateRequestPreimage(IndicatorObservationCreateParams{
				ClientTxnID: "txn-observation-create", BaseRowVersion: 7,
				SourceFieldKey: "timeline.raw_activity_text", SpanStartByte: 3, SpanEndByte: 19,
				ParsedIndicatorType: &parsedType, ResolvedIndicatorRecordID: &targetID,
			}),
			digest: observationCreateRequestHash(IndicatorObservationCreateParams{
				ClientTxnID: "txn-observation-create", BaseRowVersion: 7,
				SourceFieldKey: "timeline.raw_activity_text", SpanStartByte: 3, SpanEndByte: 19,
				ParsedIndicatorType: &parsedType, ResolvedIndicatorRecordID: &targetID,
			}),
			wantJSON: `{"client_txn_id":"txn-observation-create","base_row_version":7,"source_field_key":"timeline.raw_activity_text","span_start_byte":3,"span_end_byte":19,"parsed_indicator_type":"domain_name","resolved_indicator_record_id":"00000000-0000-4000-8000-000000000222"}`,
			wantHash: "442828a1762990721d7121d060d1c9c05dba29db36bacd4ee70947f32f119436",
		},
		{
			name: "observation resolve",
			preimage: observationResolveRequestPreimage(IndicatorObservationResolveParams{
				ClientTxnID: "txn-observation-resolve", BaseRowVersion: 8,
				ResolvedIndicatorRecordID: targetID,
			}),
			digest: observationResolveRequestHash(IndicatorObservationResolveParams{
				ClientTxnID: "txn-observation-resolve", BaseRowVersion: 8,
				ResolvedIndicatorRecordID: targetID,
			}),
			wantJSON: `{"client_txn_id":"txn-observation-resolve","base_row_version":8,"resolved_indicator_record_id":"00000000-0000-4000-8000-000000000222"}`,
			wantHash: "ed97a177a0455d4f5d2856643e838d3750479b871a12677e2fb3816da59403e6",
		},
		{
			name: "observation dismiss or restore action",
			preimage: observationActionRequestPreimage(IndicatorObservationActionParams{
				ClientTxnID: "txn-observation-action", BaseRowVersion: 9,
			}),
			digest: observationActionRequestHash(IndicatorObservationActionParams{
				ClientTxnID: "txn-observation-action", BaseRowVersion: 9,
			}),
			wantJSON: `{"client_txn_id":"txn-observation-action","base_row_version":9}`,
			wantHash: "5aa1feb7e679d13713afb0087395553481b61a353cec7f4f0a1006db9fb6e3c8",
		},
		{
			name: "lifecycle append",
			preimage: lifecycleAppendRequestPreimage(IndicatorLifecycleAppendParams{
				ClientTxnID: "txn-lifecycle", BaseRowVersion: 10,
				LifecycleState: "false_positive",
				ValidFrom:      time.Date(2026, 8, 23, 14, 15, 16, 123456000, time.UTC),
				Confidence:     &confidence, Rationale: &rationale,
				SupportRefs: []uuid.UUID{
					uuid.MustParse("00000000-0000-4000-8000-000000000333"),
					uuid.MustParse("00000000-0000-4000-8000-000000000111"),
				},
			}),
			digest: lifecycleAppendRequestHash(IndicatorLifecycleAppendParams{
				ClientTxnID: "txn-lifecycle", BaseRowVersion: 10,
				LifecycleState: "false_positive",
				ValidFrom:      time.Date(2026, 8, 23, 14, 15, 16, 123456000, time.UTC),
				Confidence:     &confidence, Rationale: &rationale,
				SupportRefs: []uuid.UUID{
					uuid.MustParse("00000000-0000-4000-8000-000000000333"),
					uuid.MustParse("00000000-0000-4000-8000-000000000111"),
				},
			}),
			wantJSON: `{"client_txn_id":"txn-lifecycle","base_row_version":10,"lifecycle_state":"false_positive","valid_from":"2026-08-23T14:15:16.123456Z","valid_to":null,"confidence":80,"rationale":"reviewed","support_refs":["00000000-0000-4000-8000-000000000333","00000000-0000-4000-8000-000000000111"],"assessor":null}`,
			wantHash: "cbdb5ffd4c7fbc099f3ca66578b39ad50d2528a82d0d227986d2b28e34f22e76",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if string(test.preimage) != test.wantJSON {
				t.Fatalf("preimage = %s, want %s", test.preimage, test.wantJSON)
			}
			wantDigest, err := hex.DecodeString(test.wantHash)
			if err != nil {
				t.Fatalf("decode golden digest: %v", err)
			}
			if !bytes.Equal(test.digest, wantDigest) {
				t.Fatalf("digest = %x, want %x", test.digest, wantDigest)
			}
		})
	}

	ordered := IndicatorLifecycleAppendParams{
		ClientTxnID: "txn-order", BaseRowVersion: 1, LifecycleState: "active",
		ValidFrom: time.Date(2026, 8, 23, 15, 0, 0, 0, time.UTC),
		SupportRefs: []uuid.UUID{
			uuid.MustParse("00000000-0000-4000-8000-000000000333"),
			uuid.MustParse("00000000-0000-4000-8000-000000000111"),
		},
	}
	reordered := ordered
	reordered.SupportRefs = append([]uuid.UUID(nil), ordered.SupportRefs...)
	reordered.SupportRefs[0], reordered.SupportRefs[1] = reordered.SupportRefs[1], reordered.SupportRefs[0]
	if bytes.Equal(lifecycleAppendRequestHash(ordered), lifecycleAppendRequestHash(reordered)) {
		t.Fatal("lifecycle replay hash must preserve caller support-reference order")
	}
	omittedObservationPreimage := observationCreateRequestPreimage(IndicatorObservationCreateParams{
		ClientTxnID: "txn-omitted-observation", BaseRowVersion: 1,
		SourceFieldKey: "timeline.raw_activity_text", SpanEndByte: 1,
	})
	if bytes.Contains(omittedObservationPreimage, []byte("parsed_indicator_type")) || bytes.Contains(omittedObservationPreimage, []byte("resolved_indicator_record_id")) {
		t.Fatalf("omitted observation members entered replay preimage: %s", omittedObservationPreimage)
	}

	store := &Store{}
	actorID := uuid.MustParse("00000000-0000-4000-8000-000000000444")
	scopeID := uuid.MustParse("00000000-0000-4000-8000-000000000555")
	dismissKey := store.childReplayKey(observationDismissRouteKey, actorID, scopeID, "txn-shared-action")
	restoreKey := store.childReplayKey(observationRestoreRouteKey, actorID, scopeID, "txn-shared-action")
	if dismissKey == restoreKey || dismissKey.RouteKey == restoreKey.RouteKey {
		t.Fatalf("dismiss and restore route keys were not isolated: dismiss=%#v restore=%#v", dismissKey, restoreKey)
	}
	otherScope := store.childReplayKey(observationDismissRouteKey, actorID, uuid.MustParse("00000000-0000-4000-8000-000000000556"), "txn-shared-action")
	if dismissKey == otherScope || dismissKey.ScopeKey == otherScope.ScopeKey {
		t.Fatalf("Indicator replay scopes were not isolated: first=%#v second=%#v", dismissKey, otherScope)
	}
}
