// Package portabilitysupport captures small source-port exports for focused assertions.
// Production and bulk tests use sealed, operation-owned member stores.
package portabilitysupport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type File struct {
	Path    string
	Payload []byte
}

type capture struct {
	files []File
	bytes int64
	paths map[string]bool
}

func Export(port sourceport.Port, ctx context.Context, input sourceport.ExportContext) ([]File, error) {
	destination := &capture{paths: map[string]bool{}}
	input.Members = destination
	if err := port.Export(ctx, input); err != nil {
		return nil, err
	}
	return destination.files, nil
}

func (c *capture) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	if err := ctx.Err(); err != nil {
		return incidentportability.Member{}, err
	}
	if c.paths[path] {
		return incidentportability.Member{}, fmt.Errorf("duplicate test member %q", path)
	}
	const maximum = 16 << 20
	payload, err := io.ReadAll(io.LimitReader(source, maximum-c.bytes+1))
	if err != nil {
		return incidentportability.Member{}, err
	}
	if int64(len(payload))+c.bytes > maximum {
		return incidentportability.Member{}, fmt.Errorf("focused export capture exceeds %d bytes", maximum)
	}
	c.bytes += int64(len(payload))
	c.paths[path] = true
	c.files = append(c.files, File{Path: path, Payload: payload})
	digest := sha256.Sum256(payload)
	return incidentportability.Member{Path: path, ByteSize: int64(len(payload)), SHA256: hex.EncodeToString(digest[:])}, nil
}

// ReadStagedAttributions reads the attribution a source port staged while
// preparing; the core flushes exactly these rows in step 9.
func ReadStagedAttributions(ctx context.Context, port sourceport.Port, prepared sourceport.Prepared, input sourceport.ImportContext) ([]incidentportability.AttributionRow, error) {
	staged, ok := port.(sourceport.StagedAttributionPort)
	if !ok {
		return nil, fmt.Errorf("source port does not stage attribution")
	}
	streams, err := staged.PreparedAttributions(prepared, input)
	if err != nil {
		return nil, err
	}
	reader := incidentportability.OpenAttributions(ctx, input.Preparation, streams...)
	var rows []incidentportability.AttributionRow
	for reader.Next() {
		rows = append(rows, reader.Row())
	}
	if err := errors.Join(reader.Err(), reader.Close()); err != nil {
		return nil, err
	}
	if reader.Count() != reader.Expected() {
		return nil, fmt.Errorf("staged attribution count %d differs from %d", reader.Count(), reader.Expected())
	}
	return rows, nil
}

// PublishFact returns input with a fact registry over its preparation store in
// the given family order, in which family has published name with keys. It
// publishes through a scoped source adapter, as a preparing family does.
func PublishFact(ctx context.Context, input sourceport.ImportContext, order []string, family, name string, keys []string) (sourceport.ImportContext, error) {
	if input.Preparation == nil {
		return input, fmt.Errorf("fact publication requires a preparation store")
	}
	input.Facts = sourceport.NewImportFacts(input.Preparation, order)
	invariant := family + ".fact_published"
	none := func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return nil }
	publisher := sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: sourceport.Descriptor{
			FamilyID: family, ContractMajor: sourceport.ContractMajor, OwnerID: "test." + family, InvariantIDs: []string{invariant},
			Paths: []sourceport.Path{{LogicalPath: "data/" + family + "-fact.ndjson", ContentRole: "source_rows", Versions: []int{input.BundleVersion}, StableIdentity: []string{"id"}, StableIdentityInvariantID: invariant}},
		},
		Export: func(context.Context, sourceport.ExportContext) error { return nil },
		Prepare: func(ctx context.Context, _ sourceport.Bundle, scoped sourceport.ImportContext) (any, error) {
			return nil, scoped.Facts.Publish(ctx, name, func(add func(string) error) error {
				for _, key := range keys {
					if err := add(key); err != nil {
						return err
					}
				}
				return nil
			})
		},
		Apply: none, Validate: none,
	})
	if _, err := publisher.PrepareImport(ctx, MapBundle{}, input); err != nil {
		return input, err
	}
	return input, nil
}

// ReferencedActors admits every source actor a bundle's rows name in a
// top-level "*_user_id" member, as a focused test's actor catalog.
func ReferencedActors(bundle MapBundle) (sourceport.ActorCatalog, error) {
	seen := map[string]bool{}
	var descriptors []sourceport.ActorDescriptor
	for _, path := range bundle.Paths() {
		if !strings.HasSuffix(path, ".ndjson") || len(bundle[path]) == 0 {
			continue
		}
		rows, err := incidentportability.DecodeNDJSON(bundle[path])
		if err != nil {
			return sourceport.ActorCatalog{}, err
		}
		for _, row := range rows {
			for column, value := range row {
				actor, ok := value.(string)
				if !strings.HasSuffix(column, "_user_id") || !ok || seen[actor] {
					continue
				}
				seen[actor] = true
				descriptors = append(descriptors, sourceport.ActorDescriptor{SourceActorID: actor, DisplayName: "Actor " + actor})
			}
		}
	}
	return sourceport.NewActorCatalog(descriptors)
}

// AttributionCapability is a focused test's final-transaction attribution
// capability. Source ports stage attribution while preparing and only require
// the capability; tests stand in for the core's step-9 flush with
// ReadStagedAttributions (Core 01 REQ-01-641).
type AttributionCapability struct{}

func (AttributionCapability) FlushImportedAttributions(context.Context, incidentportability.MemberReader, ...incidentportability.AttributionStream) error {
	return errors.New("portabilitysupport: source ports do not flush attribution")
}
