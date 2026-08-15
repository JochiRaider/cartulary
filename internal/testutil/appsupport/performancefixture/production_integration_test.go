package performancefixture

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestProductionClosedAssemblerBuildsExactDeterministicFixture_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	build := func(prefix string) fixture.Result {
		t.Helper()
		profile := profileForTest(t)
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
		pool, err := pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(pool.Close)
		if err := bootstrap.Preflight(context.Background(), bootstrap.Settings{
			ManifestPath: fixtures.Path("bootstrap-admin", "canonical.json"),
		}, pool); err != nil {
			t.Fatal(err)
		}
		actor, err := authn.NewStore(pool).GetUserByNormalizedEmail(context.Background(), "bootstrap-admin@example.test")
		if err != nil {
			t.Fatal(err)
		}
		assembler, err := NewProduction(pool, actor, prefix, profile)
		if err != nil {
			t.Fatal(err)
		}
		bundle, err := fixture.GenerateRuntimeBundle(profile, fixtureKey)
		if err != nil {
			t.Fatal(err)
		}
		state := &fixture.BuildState{
			FixtureProfileID: profile.FixtureProfileID,
			SnapshotKey:      fixtureKey,
			Seed:             profile.Seed,
			RuntimeBundle:    bundle,
		}
		result, err := assembler.Assemble(context.Background(), state)
		if err != nil {
			t.Fatal(err)
		}
		if err := fixture.ValidateReceiptRedaction(result, state); err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := build("ac043_fixture_first")
	second := build("ac043_fixture_second")
	if first.SemanticValidationDigest != second.SemanticValidationDigest {
		t.Fatalf("production builds differ: %s != %s", first.SemanticValidationDigest, second.SemanticValidationDigest)
	}
}
