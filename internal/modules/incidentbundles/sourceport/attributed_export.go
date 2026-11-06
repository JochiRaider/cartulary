package sourceport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

// AttributedRows is authored by the source owner. Query selects one JSON object
// per row, ordered by its UUID identity, with incident and exclusive-after UUID
// parameters ($1, $2). No serialized field-name inference defines actor meaning.
type AttributedRows struct {
	Path, Table, Identity, Query string
	ActorColumns                 []string
}

func ExportAttributedRows(ctx context.Context, input ExportContext, spec AttributedRows) error {
	if input.PortableAttributions == nil || input.Members == nil || input.Query == nil || spec.Identity == "" || spec.Table == "" {
		return ErrInvalidCatalog
	}
	return incidentportability.WriteMemberFrom(ctx, input.Members, spec.Path, func(ctx context.Context, output io.Writer) error {
		var after *uuid.UUID
		for {
			rows, err := input.Query.Query(ctx, spec.Query, input.IncidentID, after)
			if err != nil {
				return err
			}
			page := []map[string]any{}
			ids := []string{}
			pageBytes := 0
			for rows.Next() {
				var raw []byte
				if err := rows.Scan(&raw); err != nil {
					rows.Close()
					return err
				}
				if len(raw) > incidentportability.MaxNDJSONLogicalLineBytes {
					rows.Close()
					return ErrInvalidCatalog
				}
				var row map[string]any
				decoder := json.NewDecoder(bytes.NewReader(raw))
				decoder.UseNumber()
				if err := decoder.Decode(&row); err != nil {
					rows.Close()
					return err
				}
				value, ok := row[spec.Identity].(string)
				id, err := uuid.Parse(value)
				if !ok || err != nil || id == uuid.Nil || id.String() != value || (after != nil && value <= after.String()) {
					rows.Close()
					return fmt.Errorf("portable source identity order is invalid")
				}
				after = &id
				ids = append(ids, value)
				page = append(page, row)
				pageBytes += len(raw)
				if len(page) == 128 || pageBytes >= 1<<20 {
					break
				}
			}
			rows.Close()
			if err := rows.Err(); err != nil {
				return err
			}
			if len(page) == 0 {
				return nil
			}
			// Close source rows before querying attribution on the borrowed snapshot.
			for _, column := range spec.ActorColumns {
				actors, err := input.PortableAttributions.ResolvePortableSourceActors(ctx, input.Query, input.IncidentID, spec.Table, column, ids)
				if err != nil {
					return err
				}
				for index, row := range page {
					if actor, ok := actors[ids[index]]; ok {
						if row[column] == nil {
							return fmt.Errorf("portable attribution has no source field")
						}
						row[column] = actor
					}
					if row[column] != nil && input.Actors != nil {
						actor, ok := row[column].(string)
						if !ok {
							return ErrInvalidCatalog
						}
						actorID, err := uuid.Parse(actor)
						if err != nil || actorID == uuid.Nil {
							return ErrInvalidCatalog
						}
						if err := input.Actors.RecordActorReference(actorID); err != nil {
							return err
						}
					}
				}
			}
			for _, row := range page {
				if err := incidentportability.WriteNDJSONRecord(output, row); err != nil {
					return err
				}
			}
		}
	})
}
