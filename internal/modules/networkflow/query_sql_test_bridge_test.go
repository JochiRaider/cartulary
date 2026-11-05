package networkflow

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// Called by the authenticated pagination integration fixture using its live DB.
func AssertLiveSQLKeysetBoundaries(t *testing.T, store *store, incident uuid.UUID, table string) {
	t.Helper()
	assertOrientationSQLParity(t, store)
	for _, tc := range []struct {
		field string
		want  []int64
	}{
		{fieldBytesCount, []int64{2, 3, 1}}, {fieldEndpointAPort, []int64{1, 2, 3}},
	} {
		var position *rowCursorPosition
		var got []int64
		sort := []sortSpec{{FieldKey: tc.field, Direction: "asc"}}
		for {
			rows, more, err := store.QueryRowsPage(context.Background(), incident, []string{table}, nil, sort, position, 1)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 {
				t.Fatalf("expected one row: %v", rows)
			}
			got = append(got, rows[0].SourceRowNumber)
			next := newRowCursorPosition(rows[0], sort)
			position = &next
			if !more {
				break
			}
			if len(got) > 3 {
				t.Fatal("keyset did not advance")
			}
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("SQL %s ordering: %v want %v", tc.field, got, tc.want)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := store.QueryRowsPage(ctx, incident, []string{table}, nil, nil, nil, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("query ignored cancellation: %v", err)
	}
}

// Exercise generated parameterized predicates and cursors against PostgreSQL,
// independently of graph composition and the in-memory comparator.
func assertOrientationSQLParity(t *testing.T, store *store) {
	t.Helper()
	p0, p443 := int32(0), int32(443)
	examples := []flowRow{
		{EndpointAIP: "192.0.2.2", EndpointBIP: "198.51.100.1", EndpointAPort: nil, EndpointBPort: &p443, Direction: flowDirection{Kind: directionClientServer, FromEndpoint: "a"}},
		{EndpointAIP: "192.0.2.100", EndpointBIP: "198.51.100.1", EndpointAPort: &p0, EndpointBPort: nil, Direction: flowDirection{Kind: directionSourceDestination, FromEndpoint: "a"}},
		{EndpointAIP: "192.0.2.2", EndpointBIP: "2001:db8::1", EndpointAPort: &p443, EndpointBPort: &p0, Direction: flowDirection{Kind: directionClientServer, FromEndpoint: "b"}},
		{EndpointAIP: "192.0.2.2", EndpointBIP: "198.51.100.1", Direction: flowDirection{Kind: directionUnknown, Reason: "missing_orientation"}},
		{EndpointAIP: "192.0.2.2", EndpointBIP: "198.51.100.1", EndpointBPort: &p443, Direction: flowDirection{Kind: directionUnknown, Reason: "incomplete_orientation"}},
	}
	baseArgs := []any{}
	values := []string{}
	for i := range examples {
		row := &examples[i]
		row.RowID = fmt.Sprintf("nfr_%064d", i+1)
		row.NetworkFlowTableID = "nft_" + strings.Repeat("a", 32)
		row.SourceRowNumber = int64(i + 1)
		row.FlowStartUTC = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
		row.FlowEndUTC = row.FlowStartUTC
		parts := []string{}
		var from, reason any
		if row.Direction.FromEndpoint != "" {
			from = row.Direction.FromEndpoint
		}
		if row.Direction.Reason != "" {
			reason = row.Direction.Reason
		}
		data := []any{i, row.EndpointAIP, row.EndpointBIP, row.EndpointAPort, row.EndpointBPort, row.Direction.Kind, from, reason, row.RowID, row.NetworkFlowTableID, row.SourceRowNumber, row.FlowStartUTC, row.FlowEndUTC}
		casts := []string{"integer", "text", "text", "integer", "integer", "text", "text", "text", "text", "text", "bigint", "timestamptz", "timestamptz"}
		for j, value := range data {
			baseArgs = append(baseArgs, value)
			parts = append(parts, fmt.Sprintf("$%d::%s", len(baseArgs), casts[j]))
		}
		values = append(values, "("+strings.Join(parts, ",")+")")
	}
	cte := "WITH observations (idx,endpoint_a_ip,endpoint_b_ip,endpoint_a_port,endpoint_b_port,direction_kind,from_endpoint,direction_reason,network_flow_row_id,network_flow_table_id,source_row_number,flow_start_utc,flow_end_utc) AS (VALUES " + strings.Join(values, ",") + ") "
	filters := []queryFilter{
		{FieldKey: fieldEndpointIP, Op: "eq", Value: "192.0.2.2"},
		{FieldKey: fieldEndpointIP, Op: "in", Value: []any{"192.0.2.2", "2001:db8::1"}},
		{FieldKey: fieldEndpointIP, Op: "cidr_contains", Value: "192.0.2.0/24"},
		{FieldKey: fieldDirectionKind, Op: "eq", Value: directionUnknown},
		{FieldKey: fieldDirectionReason, Op: "in", Value: []any{"missing_orientation", "incomplete_orientation"}},
	}
	for _, field := range []string{fieldFromIP, fieldToIP, fieldEndpointAIP, fieldEndpointBIP} {
		filters = append(filters, queryFilter{FieldKey: field, Op: "eq", Value: "192.0.2.2"}, queryFilter{FieldKey: field, Op: "cidr_contains", Value: "192.0.2.0/24"})
	}
	for _, field := range []string{fieldFromPort, fieldToPort, fieldEndpointAPort, fieldEndpointBPort} {
		filters = append(filters, queryFilter{FieldKey: field, Op: "is_null"}, queryFilter{FieldKey: field, Op: "not_null"}, queryFilter{FieldKey: field, Op: "eq", Value: 0}, queryFilter{FieldKey: field, Op: "in", Value: []any{0, 443}})
	}
	for _, filter := range filters {
		args := append([]any(nil), baseArgs...)
		clause, err := appendRowFilterSQL(filter, &args)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := store.pool.Query(t.Context(), cte+"SELECT idx,coalesce(("+clause+"),false) FROM observations ORDER BY idx", args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var i int
			var actual bool
			if err := rows.Scan(&i, &actual); err != nil {
				t.Fatal(err)
			}
			expected, failure := rowMatchesFilter(examples[i], filter)
			if failure != nil || actual != expected {
				t.Fatalf("SQL/in-memory mismatch for %s/%s row %d: %v vs %v (%v)", filter.FieldKey, filter.Op, i, actual, expected, failure)
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, field := range []string{fieldFromIP, fieldToIP, fieldFromPort, fieldToPort, fieldDirectionKind, fieldDirectionReason} {
		for _, direction := range []string{"asc", "desc"} {
			specs := []sortSpec{{FieldKey: field, Direction: direction}}
			expected := append([]flowRow(nil), examples...)
			sort.Slice(expected, func(i, j int) bool {
				return compareRowToPosition(expected[i], newRowCursorPosition(expected[j], specs)) < 0
			})
			orders := []string{}
			for _, spec := range effectiveSort(specs) {
				expression, ok := rowSQLExpression(spec.FieldKey)
				if !ok {
					t.Fatal("invalid sort")
				}
				orders = append(orders, expression+" "+strings.ToUpper(spec.Direction)+" NULLS LAST")
			}
			orders = append(orders, `network_flow_table_id COLLATE "C" ASC`, `network_flow_row_id COLLATE "C" ASC`)
			for page := 0; page < len(expected); page++ {
				args := append([]any(nil), baseArgs...)
				where := "TRUE"
				if page > 0 {
					var err error
					where, err = appendRowKeysetSQL(effectiveSort(specs), newRowCursorPosition(expected[page-1], specs), &args)
					if err != nil {
						t.Fatal(err)
					}
				}
				var index int
				if err := store.pool.QueryRow(t.Context(), cte+"SELECT idx FROM observations WHERE "+where+" ORDER BY "+strings.Join(orders, ",")+" LIMIT 1", args...).Scan(&index); err != nil {
					t.Fatal(err)
				}
				if examples[index].RowID != expected[page].RowID {
					t.Fatalf("SQL cursor mismatch %s/%s page %d", field, direction, page)
				}
			}
		}
	}
}
