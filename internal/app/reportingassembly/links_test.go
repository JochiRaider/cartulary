package reportingassembly

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestLogicalSupportTargetComposition(t *testing.T) {
	t.Parallel()
	incidentID := uuid.New()
	var order []string
	provider, err := NewLinksProvider(
		logicalTargetFixture{key: "zeta", targets: map[string]string{"shared": "/evidence/shared"}, order: &order},
		logicalTargetFixture{key: "alpha", targets: map[string]string{"first": "/artifacts/first", "shared": "/evidence/shared"}, order: &order},
	)
	if err != nil {
		t.Fatalf("compose logical targets: %v", err)
	}
	targets, err := provider.collectLogicalSupportTargetsTx(context.Background(), nil, incidentID)
	if err != nil {
		t.Fatalf("collect logical targets: %v", err)
	}
	if !reflect.DeepEqual(order, []string{"alpha", "zeta"}) {
		t.Fatalf("provider order = %v, want [alpha zeta]", order)
	}
	want := map[string]string{"first": "/artifacts/first", "shared": "/evidence/shared"}
	if !reflect.DeepEqual(targets, want) {
		t.Fatalf("logical targets = %#v, want %#v", targets, want)
	}
}

func TestLogicalSupportTargetCompositionRejectsInvalidProviders(t *testing.T) {
	t.Parallel()
	if _, err := NewLinksProvider(nil); err == nil || !strings.Contains(err.Error(), "provider is required") {
		t.Fatalf("nil provider error = %v", err)
	}
	var typedNil *logicalTargetPointerFixture
	if _, err := NewLinksProvider(typedNil); err == nil || !strings.Contains(err.Error(), "provider is required") {
		t.Fatalf("typed-nil provider error = %v", err)
	}
	if _, err := NewLinksProvider(logicalTargetFixture{}); err == nil || !strings.Contains(err.Error(), "key is required") {
		t.Fatalf("empty key error = %v", err)
	}
	if _, err := NewLinksProvider(logicalTargetFixture{key: "evidence"}, logicalTargetFixture{key: "evidence"}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate provider error = %v", err)
	}
}

func TestLogicalSupportTargetCompositionRejectsConflictsAndWrapsProviderFailures(t *testing.T) {
	t.Parallel()
	provider, err := NewLinksProvider(
		logicalTargetFixture{key: "alpha", targets: map[string]string{"same": "/artifacts/same"}},
		logicalTargetFixture{key: "evidence", targets: map[string]string{"same": "/evidence/same"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.collectLogicalSupportTargetsTx(context.Background(), nil, uuid.New()); err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("conflict error = %v", err)
	}

	sentinel := errors.New("sentinel")
	provider, err = NewLinksProvider(logicalTargetFixture{key: "evidence", err: sentinel})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.collectLogicalSupportTargetsTx(context.Background(), nil, uuid.New()); !errors.Is(err, sentinel) || !strings.Contains(err.Error(), "evidence") {
		t.Fatalf("provider failure = %v", err)
	}
}

type logicalTargetFixture struct {
	key     string
	targets map[string]string
	err     error
	order   *[]string
}

type logicalTargetPointerFixture struct{}

func (*logicalTargetPointerFixture) ProviderKey() string { return "pointer" }

func (*logicalTargetPointerFixture) CollectLogicalSupportTargetsTx(context.Context, pgx.Tx, uuid.UUID) (map[string]string, error) {
	return nil, nil
}

func (fixture logicalTargetFixture) ProviderKey() string { return fixture.key }

func (fixture logicalTargetFixture) CollectLogicalSupportTargetsTx(context.Context, pgx.Tx, uuid.UUID) (map[string]string, error) {
	if fixture.order != nil {
		*fixture.order = append(*fixture.order, fixture.key)
	}
	return fixture.targets, fixture.err
}
