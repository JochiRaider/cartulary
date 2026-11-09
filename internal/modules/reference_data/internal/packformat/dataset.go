package packformat

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

func datasetExpected(m Manifest) (dataset.Expected, error) {
	if m.ProfileID != dataset.ProfileID || m.ProfileVersion != "1" || m.Summary.Kind != "objects_relationships" || m.Summary.Objects == nil || m.Summary.Relationships == nil || m.Summary.Entries != nil {
		return dataset.Expected{}, fail("contract_incompatible")
	}
	return dataset.Expected{Objects: *m.Summary.Objects, Relationships: *m.Summary.Relationships, SourceArtifacts: len(m.Artifacts)}, nil
}

func datasetError(err error) error {
	var failure *dataset.Failure
	if errors.As(err, &failure) {
		f := &Failure{Code: failure.Code}
		if failure.Limit != "" {
			f.Details = unsignedLimitFinding("$", failure.Limit, failure.Maximum, failure.Observed).Details
		}
		return f
	}
	if errors.Is(err, dataset.ErrStorage) {
		return ErrStorage
	}
	return err
}

func datasetFindings(emit FindingSink) schema.FindingSink {
	return func(f schema.Finding) error {
		return emit(Finding{Path: f.Path, EntryID: f.EntryID, Details: SafeDetails{LimitID: f.Details.LimitID, ExpectedToken: f.Details.ExpectedToken, ActualToken: f.Details.ActualToken, RelatedPackKey: f.Details.RelatedPackKey, RelatedPackVersion: f.Details.RelatedPackVersion}})
	}
}

type datasetIndex struct{ sink ContentSink }

func (s datasetIndex) Append(ctx context.Context, row dataset.Row) error {
	keys := make([]LookupKey, 0, len(row.Postings))
	for _, p := range row.Postings {
		keys = append(keys, LookupKey{Kind: p.Lookup, Value: p.Key, Order: row.ID})
	}
	return s.sink.Append(ctx, ContentRow{ID: row.ID, Shape: row.Shape, Canonical: row.Canonical, Keys: keys})
}

func validateDataset(ctx context.Context, m Manifest, source ContentSource, sink ContentSink) error {
	expected, err := datasetExpected(m)
	if err != nil {
		return err
	}
	var index dataset.Sink
	if sink != nil {
		index = datasetIndex{sink}
	}
	_, err = dataset.Validate(ctx, source, expected, index)
	return datasetError(err)
}
