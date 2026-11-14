package performancefixturelifecycle

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSuiteCleanupPreservesIndependentFailures(t *testing.T) {
	t.Parallel()
	const suiteID = "cleanup-suite"
	allOwned := make([]ownedTemplate, 0, 9)
	for _, digit := range []string{"a", "b", "c", "d", "e", "f", "0", "1", "2"} {
		key := strings.Repeat(digit, 64)
		allOwned = append(allOwned, ownedTemplate{name: templateName(suiteID, key), owner: templateOwner(suiteID, key)})
	}
	owned := allOwned[:3]
	t.Run("inventory settlement", func(t *testing.T) {
		scanFailure, iterationFailure, closeFailure := errors.New("scan"), errors.New("iteration"), errors.New("close")
		for _, failScan := range []bool{false, true} {
			rows := &cleanupInventoryRows{values: owned, iterationErr: iterationFailure, closeErr: closeFailure}
			if failScan {
				rows.scanErr = scanFailure
			}
			_, err := readOwnedTemplates(rows)
			if !rows.closed || !errors.Is(err, iterationFailure) || !errors.Is(err, closeFailure) || failScan && !errors.Is(err, scanFailure) {
				t.Fatal("incomplete inventory lost a failure or remained open")
			}
		}
		rows := &cleanupInventoryRows{values: owned}
		actual, err := readOwnedTemplates(rows)
		if err != nil || !rows.closed || !reflect.DeepEqual(actual, owned) {
			t.Fatal("complete inventory changed or remained open")
		}
	})
	t.Run("admission precedes destruction", func(t *testing.T) {
		for name, inventory := range map[string][]ownedTemplate{
			"foreign owner":        {owned[0], {name: owned[1].name, owner: templateOwner("other-suite", strings.Repeat("b", 64))}},
			"missing owner prefix": {owned[0], {name: owned[1].name, owner: strings.Repeat("b", 64)}},
			"foreign name":         {owned[0], {name: "ct_other_template", owner: owned[1].owner}},
			"duplicate":            {owned[0], owned[0]},
			"oversized":            allOwned,
		} {
			t.Run(name, func(t *testing.T) {
				calls := 0
				err := retireOwnedTemplates(context.Background(), suiteID, inventory, []string{"private-root"}, func(context.Context, string) error { calls++; return nil }, func(string) error { calls++; return nil })
				if err == nil || calls != 0 {
					t.Fatal("invalid inventory authorized destructive work")
				}
			})
		}
	})
	t.Run("independent retirement and proof retention", func(t *testing.T) {
		firstFailure, lastFailure := errors.New("first retirement"), errors.New("last retirement")
		var attempted, removed []string
		roots := []string{"template-proof", "clone-proof"}
		err := retireOwnedTemplates(context.Background(), suiteID, owned, roots, func(_ context.Context, name string) error {
			attempted = append(attempted, name)
			if name == owned[0].name {
				return firstFailure
			}
			if name == owned[2].name {
				return lastFailure
			}
			return nil
		}, func(root string) error { removed = append(removed, root); return nil })
		want := []string{owned[0].name, owned[1].name, owned[2].name}
		if !errors.Is(err, firstFailure) || !errors.Is(err, lastFailure) || !reflect.DeepEqual(attempted, want) || len(removed) != 0 {
			t.Fatal("independent retirement stopped, lost causes or removed unresolved proof")
		}
		// A later exact attempt may settle the same targets, including an already
		// absent target, before both private roots can be retired.
		attempted = nil
		removeFailure := errors.New("proof removal")
		err = retireOwnedTemplates(context.Background(), suiteID, owned, roots, func(_ context.Context, name string) error {
			attempted = append(attempted, name)
			return nil
		}, func(root string) error {
			removed = append(removed, root)
			if root == roots[0] {
				return removeFailure
			}
			return nil
		})
		if !errors.Is(err, removeFailure) || !reflect.DeepEqual(attempted, want) || !reflect.DeepEqual(removed, roots) {
			t.Fatal("settled retirement skipped independent proof cleanup or lost its failure")
		}
	})
}

type cleanupInventoryRows struct {
	values       []ownedTemplate
	index        int
	scanErr      error
	iterationErr error
	closeErr     error
	closed       bool
}

func (rows *cleanupInventoryRows) Next() bool { return rows.index < len(rows.values) }
func (rows *cleanupInventoryRows) Scan(values ...any) error {
	if rows.scanErr != nil {
		return rows.scanErr
	}
	value := rows.values[rows.index]
	rows.index++
	*values[0].(*string), *values[1].(*string) = value.name, value.owner
	return nil
}
func (rows *cleanupInventoryRows) Err() error { return rows.iterationErr }
func (rows *cleanupInventoryRows) Close() error {
	rows.closed = true
	return rows.closeErr
}
