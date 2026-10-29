package storetest_test

import (
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSeedIncidentMembershipSQLCanonicalIdentity(t *testing.T) {
	db := pgtest.Start(t).OpenIsolatedDatabaseT(t, "incident-sql-fixture", postgres.PurposeRecovery)
	actor := authflowtest.SeedActor(t, db, authflowtest.ActorAttributes{Email: "incident-sql-fixture@example.test", DisplayName: "Fixture actor"})
	for _, tag := range []string{"case_distinguishing", "  cafe\u0301"} {
		t.Run(tag, func(t *testing.T) {
			id := storetest.SeedIncidentMembershipSQL(t, db, actor.ID.String(), tag)
			var key, canonical, role, member, addedBy string
			if err := db.QueryRowContext(t.Context(), `
SELECT i.incident_key, i.incident_key_canonical, m.role, m.user_id::text, m.added_by_user_id::text
FROM incidents i JOIN incident_memberships m ON m.incident_id = i.id
WHERE i.id = $1`, id).Scan(&key, &canonical, &role, &member, &addedBy); err != nil {
				t.Fatal(err)
			}
			if key != strings.TrimSpace(key) || !norm.NFC.IsNormalString(key) || canonical != key {
				t.Fatalf("fixture identity is not canonical: key=%q canonical=%q", key, canonical)
			}
			if strings.ToLower(key) == key {
				t.Fatal("fixture must distinguish case folding from canonicalization")
			}
			if role != "admin" || member != actor.ID.String() || addedBy != actor.ID.String() {
				t.Fatalf("incorrect bootstrap membership: role=%s member=%s added_by=%s", role, member, addedBy)
			}
		})
	}
}
