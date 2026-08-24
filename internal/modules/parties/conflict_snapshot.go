package parties

import conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"

func newPartyConflictSnapshotProjector() (conflicttokens.RevisionSnapshotProjector, error) {
	projector, err := conflicttokens.NewRevisionSnapshotProjector(
		"cartulary.revisions.snapshot.party.v1",
		map[string]string{
			"party.display_name":      "display_name",
			"party.party_kind":        "party_kind",
			"party.organization_name": "organization_name",
			"party.role_title":        "role_title",
			"party.primary_email":     "primary_email",
			"party.timezone_name":     "timezone_name",
			"party.external_ref":      "external_ref",
			"party.notes":             "notes",
		},
	)
	if err != nil {
		return conflicttokens.RevisionSnapshotProjector{}, err
	}
	return projector, nil
}
