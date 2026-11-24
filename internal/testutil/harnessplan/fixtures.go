package harnessplan

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	"github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

// rowFixture is a populated fixture row's planned snapshot: the profile and
// predicate the row measures, the snapshot key, and the builder and
// PostgreSQL cluster every row of that snapshot shares.
type rowFixture struct {
	profile   performancefixtureprofile.Profile
	predicate string
	key       string
	builder   FixtureBuilder
	builderID string
	// cluster holds the sealed snapshot from its build until its last clone.
	cluster string
}

// resolveFixtures routes every catalog row to its populated fixture through
// the fixture owner and keys each snapshot. Routing is checked for every row
// before any plan exists: a populated profile binds a browser row to exactly
// one predicate, its partition declares the same profile, and each active
// profile has exactly one builder.
func (in *Inputs) resolveFixtures() error {
	fail := func(source, format string, args ...any) error {
		return fmt.Errorf("%w: %s: %s", ErrInput, source, fmt.Sprintf(format, args...))
	}
	builders := map[string]FixtureBuilder{}
	for i, builder := range in.Fixtures.Builders {
		if i > 0 && in.Fixtures.Builders[i-1].FixtureProfileID >= builder.FixtureProfileID {
			return fail(fixtureBuilderPath, "builders must be sorted by fixture profile without duplicates")
		}
		builders[builder.FixtureProfileID] = builder
	}
	for _, profile := range performancefixtureprofile.Profiles() {
		if _, ok := builders[profile.FixtureProfileID]; ok != (profile.Status == "active") {
			return fail(fixtureBuilderPath, "builders must cover exactly the active fixture profiles (%s)", profile.FixtureProfileID)
		}
		delete(builders, profile.FixtureProfileID)
	}
	if len(builders) != 0 {
		return fail(fixtureBuilderPath, "builders name unknown fixture profiles")
	}
	for _, builder := range in.Fixtures.Builders {
		builders[builder.FixtureProfileID] = builder
	}
	in.rowFixtures = map[string]*rowFixture{}
	for _, row := range in.Catalog.Rows {
		fixture, ok, err := performancefixture.ResolveRow(row.FixtureProfileID, row.VerificationIDs)
		if err != nil {
			return fail(catalogRegistryPath, "catalog row %s: %v", row.RowID, err)
		}
		if !ok {
			continue
		}
		if row.Runner != RunnerPlaywright || len(fixture.Predicates) != 1 {
			return fail(catalogRegistryPath, "catalog row %s: a populated fixture binds one browser row to exactly one predicate", row.RowID)
		}
		key, err := performancefixture.SnapshotKey(fixture.Profile, in.MigrationDigest)
		if err != nil {
			return fail(catalogRegistryPath, "catalog row %s: %v", row.RowID, err)
		}
		builder := builders[row.FixtureProfileID]
		builderID := "fixture_snapshot:" + builder.RuntimeProfileID + ":" + row.FixtureProfileID + ":" + key
		in.rowFixtures[row.RowID] = &rowFixture{profile: fixture.Profile, predicate: fixture.Predicates[0], key: key,
			builder: builder, builderID: builderID, cluster: "postgres:measurement:" + digest(builderID)}
	}
	for _, partition := range in.Browser.Partitions {
		for _, id := range partition.SelectedRowIDs {
			profile := ""
			if fixture := in.rowFixtures[id]; fixture != nil {
				profile = fixture.profile.FixtureProfileID
			}
			if profile != partition.FixtureProfileID {
				return fail(browserBatchPath, "partition %s declares fixture profile %q but row %s has %q", partition.Name, partition.FixtureProfileID, id, profile)
			}
		}
	}
	return nil
}
