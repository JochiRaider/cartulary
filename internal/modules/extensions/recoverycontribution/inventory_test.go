package recoverycontribution

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestStagedObjectRecoveryInventory_Unit(t *testing.T) {
	ctx := t.Context()
	body := []byte("staged bytes")
	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	rows := &inventoryRows{values: []inventoryRow{
		{"first", "storage/first", int64(len(body)), digest},
		{"second", "storage/second", 0, hex.EncodeToString(sha256.New().Sum(nil))},
	}}
	snapshot := &inventorySnapshot{rows: rows}
	source := &inventoryBytes{bodies: map[string][]byte{"storage/first": body, "storage/second": {}}}
	provider := VNextRecoveryObjectInventory(source)
	if provider.OwnerID() != "module.extensions" || provider.ObjectFamilyID() != "extensions.staged_objects" || provider.InventoryAlgorithmID() != "extensions.snapshot_staged_object_inventory.v1" {
		t.Fatal("source-owner inventory identity changed")
	}
	members, err := provider.Inventory(ctx, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.queries != 1 || rows.closes != 1 || len(members) != 2 || len(source.opened) != 0 {
		t.Fatalf("inventory query/closure/member/deferred-open counts = %d/%d/%d/%d", snapshot.queries, rows.closes, len(members), len(source.opened))
	}
	for i, member := range members {
		want := rows.values[i]
		if member.LogicalObjectID != want.id || member.StorageKey != want.key || member.PlaintextBytes != want.size || member.PlaintextSHA256 != want.digest || member.ContentType != "application/octet-stream" {
			t.Fatalf("member metadata differs: %#v", member)
		}
		reader, err := member.Open(ctx)
		if err != nil {
			t.Fatal(err)
		}
		got, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(got, source.bodies[want.key]) {
			t.Fatalf("member bytes: %v / %v", readErr, closeErr)
		}
	}
	if !reflect.DeepEqual(source.opened, []string{"storage/first", "storage/second"}) {
		t.Fatalf("deferred opens used wrong identities: %v", source.opened)
	}
	if source.stats != 0 {
		t.Fatal("enumeration unexpectedly accessed object metadata")
	}

	t.Run("empty", func(t *testing.T) {
		rows := &inventoryRows{}
		members, err := VNextRecoveryObjectInventory(nil).Inventory(ctx, &inventorySnapshot{rows: rows})
		if err != nil || len(members) != 0 || rows.closes != 1 {
			t.Fatalf("empty inventory: %v, closes=%d", err, rows.closes)
		}
	})
	t.Run("deferred_open_failure", func(t *testing.T) {
		failure := errors.New("object unavailable")
		source := &inventoryBytes{failure: failure}
		members, err := VNextRecoveryObjectInventory(source).Inventory(ctx, &inventorySnapshot{rows: &inventoryRows{values: rows.values[:1]}})
		if err != nil || len(source.opened) != 0 {
			t.Fatalf("enumeration opened bytes: %v", err)
		}
		if _, err := members[0].Open(ctx); !errors.Is(err, failure) {
			t.Fatalf("deferred error = %v", err)
		}
	})
}

func TestStagedObjectRecoveryInventoryFailures_Unit(t *testing.T) {
	failure := errors.New("snapshot read failed")
	for _, stage := range []string{"query", "scan", "iteration", "cancellation"} {
		t.Run(stage, func(t *testing.T) {
			rows := &inventoryRows{values: []inventoryRow{{id: "first"}, {id: "second"}}}
			snapshot := &inventorySnapshot{rows: rows}
			wantErr, wantCloses := failure, 1
			switch stage {
			case "query":
				snapshot.queryErr, wantCloses = failure, 0
			case "scan":
				rows.scanErr = failure // Fail after one member has been collected.
			case "iteration":
				rows.iterationErr = failure
			case "cancellation":
				snapshot.queryErr, wantErr, wantCloses = context.Canceled, context.Canceled, 0
			}
			members, err := VNextRecoveryObjectInventory(nil).Inventory(t.Context(), snapshot)
			if !errors.Is(err, wantErr) || members != nil || rows.closes != wantCloses {
				t.Fatalf("error=%v, partial inventory=%v, closes=%d; want wrapped error, nil inventory, %d closes", err, members, rows.closes, wantCloses)
			}
		})
	}
}

type inventoryRow struct {
	id, key string
	size    int64
	digest  string
}
type inventoryRows struct {
	values                []inventoryRow
	next, closes          int
	scanErr, iterationErr error
}

func (r *inventoryRows) Next() bool {
	if r.next >= len(r.values) {
		return false
	}
	r.next++
	return true
}
func (r *inventoryRows) Scan(dest ...any) error {
	if r.scanErr != nil && r.next == 2 {
		return r.scanErr
	}
	v := r.values[r.next-1]
	*dest[0].(*string), *dest[1].(*string), *dest[2].(*int64), *dest[3].(*string) = v.id, v.key, v.size, v.digest
	return nil
}
func (r *inventoryRows) Err() error { return r.iterationErr }
func (r *inventoryRows) Close()     { r.closes++ }

type inventorySnapshot struct {
	rows     *inventoryRows
	queryErr error
	queries  int
}

func (s *inventorySnapshot) QueryRows(context.Context, string, ...any) (recovery.VNextRows, error) {
	s.queries++
	return s.rows, s.queryErr
}
func (*inventorySnapshot) StreamCanonicalTableRows(context.Context, string, func(json.RawMessage) error) error {
	return errors.New("unexpected table stream")
}

type inventoryBytes struct {
	bodies  map[string][]byte
	opened  []string
	stats   int
	failure error
}

func (s *inventoryBytes) OpenRecoveryObject(_ context.Context, key string) (io.ReadCloser, error) {
	s.opened = append(s.opened, key)
	if s.failure != nil {
		return nil, s.failure
	}
	body, ok := s.bodies[key]
	if !ok {
		return nil, errors.New("unexpected object key")
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}
func (s *inventoryBytes) StatRecoveryObject(context.Context, string) (recovery.VNextObjectSourceInfo, error) {
	s.stats++
	return recovery.VNextObjectSourceInfo{}, errors.New("unexpected object stat")
}
