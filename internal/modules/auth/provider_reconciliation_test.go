package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestEnterpriseProviderReconciliation_Integration(t *testing.T) {
	for _, tc := range []struct {
		name string
		keys []string
	}{
		{"shrinking", []string{"a-kept"}},
		{"equal size", []string{"m-new", "a-kept"}},
		{"growing", []string{"n-new", "m-new", "a-kept"}},
		{"empty", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
			fixture := newStoreFixture(t, "auth-provider-reconciliation", now)
			ctx, store := t.Context(), fixture.store
			definition := func(key string) authn.EnterpriseAuthProviderDefinition {
				return authn.EnterpriseAuthProviderDefinition{
					ProviderKey: key, ProviderType: "oidc", DisplayName: "Provider", Enabled: true,
				}
			}
			if err := store.ReconcileEnterpriseAuthProviders(ctx, []authn.EnterpriseAuthProviderDefinition{definition("z-removed"), definition("a-kept")}, now); err != nil {
				t.Fatal(err)
			}
			removed, err := store.GetEnterpriseAuthProviderByKey(ctx, "z-removed")
			if err != nil {
				t.Fatal(err)
			}
			kept, err := store.GetEnterpriseAuthProviderByKey(ctx, "a-kept")
			if err != nil {
				t.Fatal(err)
			}
			user := storetest.SeedActor(t, fixture.harness.DB, storetest.ActorAttributes{Email: "linked@example.test", DisplayName: "Linked"})
			originalUser, err := store.GetUserByID(ctx, user.ID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := fixture.harness.DB.Exec(ctx, `
INSERT INTO enterprise_auth_bindings (user_id, provider_id, provider_key, provider_type, provider_subject, created_by_user_id)
VALUES ($1, $2, 'z-removed', 'oidc', 'retained-subject', $1)
`, user.ID, removed.ID); err != nil {
				t.Fatal(err)
			}
			bindingSnapshot := func() string {
				t.Helper()
				var snapshot string
				if err := fixture.harness.DB.QueryRow(ctx, `SELECT row_to_json(b)::text FROM enterprise_auth_bindings b WHERE user_id=$1`, user.ID).Scan(&snapshot); err != nil {
					t.Fatal(err)
				}
				return snapshot
			}
			originalBinding := bindingSnapshot()
			definitions := make([]authn.EnterpriseAuthProviderDefinition, 0, len(tc.keys))
			for _, key := range tc.keys {
				definitions = append(definitions, definition(key))
			}
			changedAt := now.Add(time.Minute)
			if err := store.ReconcileEnterpriseAuthProviders(ctx, definitions, changedAt); err != nil {
				t.Fatal(err)
			}
			for _, original := range []authn.EnterpriseAuthProviderRecord{kept, removed} {
				got, err := store.GetEnterpriseAuthProviderByKey(ctx, original.ProviderKey)
				if err != nil {
					t.Fatal(err)
				}
				wantEnabled := slices.Contains(tc.keys, original.ProviderKey)
				if got.ID != original.ID || !got.CreatedAt.Equal(original.CreatedAt) || got.IsEnabled != wantEnabled || got.IsInteractive != wantEnabled || !got.UpdatedAt.Equal(changedAt) {
					t.Fatalf("provider %s identity, admission or timestamp changed incorrectly", original.ProviderKey)
				}
			}
			fixture.service.enterpriseAdmitted = true
			discovery := httptest.NewRecorder()
			fixture.service.handleEnterpriseProviders(discovery, httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil))
			if discovery.Code != http.StatusOK {
				t.Fatalf("discovery status = %d", discovery.Code)
			}
			providers := decodeSuccessData(t, discovery)["providers"].([]any)
			wantKeys := append([]string(nil), tc.keys...)
			slices.Sort(wantKeys)
			if len(providers) != len(wantKeys) {
				t.Fatalf("discovery has %d providers, want %d", len(providers), len(wantKeys))
			}
			for i, key := range wantKeys {
				want := map[string]any{"provider_key": key, "provider_type": "oidc", "display_name": "Provider"}
				if !reflect.DeepEqual(providers[i], want) {
					t.Fatalf("discovery provider %d differs from the configured set", i)
				}
			}
			begin := httptest.NewRecorder()
			fixture.service.handleEnterpriseProviders(begin, newJSONRequest(t, http.MethodPost, "/api/v1/auth/providers/z-removed/begin", `{}`))
			requireErrorEnvelope(t, begin, http.StatusNotFound, "auth_provider_disabled", "")
			if bindingSnapshot() != originalBinding {
				t.Fatal("provider reconciliation changed binding identity or lineage")
			}
			gotUser, err := store.GetUserByID(ctx, user.ID)
			if err != nil || !reflect.DeepEqual(gotUser, originalUser) {
				t.Fatal("provider reconciliation changed user credentials or authority")
			}

			// An earlier mutable update/new insert must roll back when a later key
			// attempts a forbidden type change. Reversing input also checks sorting.
			before, err := store.GetEnterpriseAuthProviderByKey(ctx, "a-kept")
			if err != nil {
				t.Fatal(err)
			}
			changed := definition("a-kept")
			changed.DisplayName = "Must roll back"
			invalid := definition("z-removed")
			invalid.ProviderType = "saml"
			err = store.ReconcileEnterpriseAuthProviders(ctx, []authn.EnterpriseAuthProviderDefinition{invalid, definition("b-rollback"), changed}, changedAt.Add(time.Minute))
			if !errors.Is(err, authn.ErrAuthProviderTypeChangeNotSupported) {
				t.Fatalf("type change returned %v", err)
			}
			after, err := store.GetEnterpriseAuthProviderByKey(ctx, "a-kept")
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("failed reconciliation exposed a mutable update")
			}
			if _, err := store.GetEnterpriseAuthProviderByKey(ctx, "b-rollback"); !errors.Is(err, authn.ErrAuthProviderNotFound) {
				t.Fatal("failed reconciliation exposed a new provider")
			}
			if bindingSnapshot() != originalBinding {
				t.Fatal("failed reconciliation changed a binding")
			}

			// Explicit disable and later re-enable are configuration updates that
			// preserve provider identity and the linked user binding.
			disabled := definition("z-removed")
			disabled.Enabled = false
			disabled.DisplayName = "Renamed"
			for _, enabled := range []bool{false, true} {
				disabled.Enabled = enabled
				if err := store.ReconcileEnterpriseAuthProviders(ctx, []authn.EnterpriseAuthProviderDefinition{disabled}, changedAt); err != nil {
					t.Fatal(err)
				}
				got, err := store.GetEnterpriseAuthProviderByKey(ctx, "z-removed")
				if err != nil || got.ID != removed.ID || got.DisplayName != "Renamed" || got.IsEnabled != enabled || got.IsInteractive != enabled || bindingSnapshot() != originalBinding {
					t.Fatal("disable/re-enable failed to preserve provider identity and binding")
				}
			}
		})
	}
}
