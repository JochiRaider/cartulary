package producer

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
)

const PackKey = "enrichment.entra_app_ids"
const ReleaseRepository = "cartulary.entra_app_ids"
const DatasetFixtureRepository = "fixture.entra_app_ids.dataset.v1"

func publisherScope(repository string) (publication.Scope, error) {
	switch repository {
	case ReleaseRepository:
		return publication.ReleaseScope(ReleaseRepository, PackKey)
	case DatasetFixtureRepository:
		return publication.FixtureScope(DatasetFixtureRepository, PackKey)
	default:
		return publication.Scope{}, errors.New("publisher identity outside this producer")
	}
}
func finishLogical(ctx context.Context, m envelope.Manifest, members packageSource, ledger, output string) (publication.LogicalReport, error) {
	scope, err := publisherScope(*m.Repository)
	if err != nil {
		return publication.LogicalReport{}, err
	}
	return publication.PublishLogical(ctx, publication.LogicalInput{Scope: scope, Manifest: m, Members: members, LedgerRoot: ledger, OutputRoot: output})
}
