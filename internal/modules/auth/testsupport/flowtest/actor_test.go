package flowtest_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestActorFixturesSharePreparationButOwnIndependentUsers(t *testing.T) {
	// Two real connection APIs observe committed rows. This fixture deliberately
	// owns a dedicated database rather than borrowing a rollback transaction.
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "actor-fixtures")
	config, err := pgx.ParseConfig(database.DSN)
	if err != nil {
		t.Fatal("parse actor fixture connection")
	}
	sqlDB := stdlib.OpenDB(*config)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error("close actor fixture SQL connection")
		}
	})
	pool, err := pgxpool.New(t.Context(), database.DSN)
	if err != nil {
		t.Fatal("open actor fixture store connection")
	}
	t.Cleanup(pool.Close)
	first := flowtest.SeedActor(t, sqlDB, flowtest.ActorAttributes{
		Email: "first-actor@example.test", DisplayName: "First actor",
	})
	secondID := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	second := storetest.SeedActor(t, pool, storetest.ActorAttributes{
		ID: secondID, Email: "second-actor@example.test", DisplayName: "Second actor", Disabled: true, DeploymentAdmin: true,
	})
	if first.ID == second.ID || second.ID != secondID || !first.IsActive || second.IsActive || !second.IsDeploymentAdmin ||
		first.PasswordHash != "" || second.PasswordHash != "" {
		t.Fatal("actor attributes or independent identity were not preserved")
	}
	readHash := func(user authn.UserRecord) string {
		t.Helper()
		var hash string
		if err := pool.QueryRow(t.Context(), "SELECT password_hash FROM users WHERE id = $1", user.ID).Scan(&hash); err != nil {
			t.Fatal("read actor fixture credential binding")
		}
		return hash
	}
	original := readHash(first)
	if original == "" || readHash(second) != original {
		t.Fatal("actor APIs did not share the prepared immutable input")
	}
	if _, err := sqlDB.ExecContext(t.Context(), "UPDATE users SET password_hash = 'invalidated-test-credential', display_name = 'Changed', is_active = false WHERE id = $1", first.ID); err != nil {
		t.Fatal("mutate first actor independently")
	}
	thirdID := uuid.MustParse("10000000-0000-4000-8000-000000000003")
	third := flowtest.SeedActor(t, sqlDB, flowtest.ActorAttributes{
		ID: thirdID, Email: "third-actor@example.test", DisplayName: "Third actor",
	})
	if readHash(second) != original || readHash(third) != original || third.ID != thirdID || third.ID == first.ID || third.ID == second.ID {
		t.Fatal("an actor mutation affected a sibling or later fixture")
	}
	var displayName string
	if err := pool.QueryRow(t.Context(), "SELECT display_name FROM users WHERE id = $1", second.ID).Scan(&displayName); err != nil || displayName != "Second actor" {
		t.Fatal("actor attributes leaked between users")
	}
	// Credential-sensitive fixtures still derive and verify the supplied password.
	const password = "Independent credential workflow 1!"
	credentialUser := storetest.SeedLocalUserRecord(t, pool, "credential-user@example.test", "Credential user", password, false, false, true)
	valid, err := authn.VerifyPasswordHash(context.Background(), credentialUser.PasswordHash, password)
	if err != nil || !valid || credentialUser.PasswordHash == original {
		t.Fatal("password-taking fixture did not preserve the real credential workflow")
	}
}
