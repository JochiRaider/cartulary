package appsupport

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestApplicationFixtureAdmission(t *testing.T) {
	t.Run("runtime needs no service assignment", func(t *testing.T) {
		runtime := StartRuntime(t)
		if _, err := os.Stat(runtime.referencePackRoot); err != nil {
			t.Fatal("local deployment root", err)
		}
	})
	for _, tc := range []struct {
		name            string
		database, store bool
		declaration     string
	}{
		{"neither supplied", false, false, "object_store,postgres"},
		{"database supplied", true, false, "object_store,postgres"},
		{"store supplied", false, true, "postgres"},
		{"both supplied", true, true, "postgres"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			database := &pgtest.TestDatabase{Name: "borrowed"}
			store := &admissionStore{}
			options := ServerOptions{Prefix: "fixture", TestRouteMode: httptestx.TestRouteModeDisabled}
			if tc.database {
				options.Database = database
			}
			if tc.store {
				options.ObjectStore = store
			}
			databaseStarts, storeStarts := 0, 0
			starters := fixtureStarters{
				database: func(testing.TB, string) *pgtest.TestDatabase {
					databaseStarts++
					return database
				},
				objectStore: func(testing.TB, string) (objectstore.Store, map[string]string) {
					storeStarts++
					return store, map[string]string{"fixture": "owned"}
				},
			}
			result, err := acquireServerFixtures(t, options, map[string]string{
				suiteservices.HarnessServiceDependenciesEnv: tc.declaration,
			}, starters)
			if err != nil {
				t.Fatal(err)
			}
			if result.database != database || result.store != store {
				t.Fatal("fixture identity was replaced")
			}
			wantDatabases, wantStores := 0, 0
			if !tc.database {
				wantDatabases = 1
			}
			if !tc.store {
				wantStores = 1
			}
			if databaseStarts != wantDatabases || storeStarts != wantStores {
				t.Fatalf("unnecessary or missing allocation: database=%d store=%d", databaseStarts, storeStarts)
			}
			if tc.store && result.objectEnv != nil {
				t.Fatal("borrowed store synthesized managed-service settings")
			}
			if !tc.store && result.objectEnv["fixture"] != "owned" {
				t.Fatal("owned store lost its settings")
			}
			// Even a supplied database needs PostgreSQL runtime admission. A supplied
			// store removes only this helper's object-store requirement.
			for _, missing := range strings.Split(tc.declaration, ",") {
				remaining := []string{}
				for _, service := range strings.Split(tc.declaration, ",") {
					if service != missing {
						remaining = append(remaining, service)
					}
				}
				databaseStarts, storeStarts = 0, 0
				options.Env = map[string]string{suiteservices.HarnessServiceDependenciesEnv: "object_store,postgres"}
				_, err = acquireServerFixtures(t, options, map[string]string{
					suiteservices.HarnessServiceDependenciesEnv: strings.Join(remaining, ","),
				}, starters)
				var dependencyErr *suiteservices.ServiceDependencyError
				if !errors.As(err, &dependencyErr) || dependencyErr.Service != missing {
					t.Fatalf("missing %s admission: %v", missing, err)
				}
				if databaseStarts != 0 || storeStarts != 0 {
					t.Fatal("acquisition preceded complete dependency admission")
				}
				envelope := runtimeDependencyEnvelope(err)
				if envelope.Validate() != nil || envelope.FailureClass != "harness" ||
					envelope.FailureReason != "fixture_error" || envelope.Service != missing ||
					envelope.ReadinessStage != "dependency_guard" {
					t.Fatalf("incorrect dependency failure classification: %#v", envelope)
				}
			}
		})
	}
	t.Run("invalid options reject before acquisition", func(t *testing.T) {
		for _, options := range []ServerOptions{
			{TestRouteMode: httptestx.TestRouteModeDisabled},
			{Prefix: "fixture"},
			{Prefix: "fixture", TestRouteMode: "unknown"},
		} {
			// Nil starters make any attempted acquisition a failure of this boundary.
			if _, err := acquireServerFixtures(t, options, nil, fixtureStarters{}); err == nil {
				t.Fatal("invalid options admitted")
			}
		}
	})
}

type admissionStore struct{ objectstore.Store }
