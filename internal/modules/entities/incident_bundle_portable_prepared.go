package entities

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

// Staged Entities streams (Core 01 REQ-01-641). Preparation admits every row
// and derives its identifier claims once; the final transaction reads these
// streams and checks only database-side facts. Paths are relative to the port's
// preparation scope.
const (
	preparedHostsPath       = "hosts.rows"
	preparedIdentitiesPath  = "identities.rows"
	preparedMentionsPath    = "entity_mentions.rows"
	preparedPreservedPath   = "entity_preserved_identifiers.rows"
	preparedAliasesPath     = "entity_aliases.rows"
	preparedEntityAttribute = "attribution.rows"
	// entitySetBatch bounds one set-based statement's parameter arrays.
	entitySetBatch = 512
)

// entityClaimCandidate is one identifier claim derived while preparing. An
// invalid normalization is reported only when the transaction finds the
// owning record active.
type entityClaimCandidate struct {
	Identifier string
	Normalized string
	Valid      bool
}

func entityClaim(identifier string, raw *string) (entityClaimCandidate, bool) {
	if raw == nil {
		return entityClaimCandidate{}, false
	}
	normalized, ok := fieldnorm.NormalizeIdentifier(identifier, *raw)
	return entityClaimCandidate{Identifier: identifier, Normalized: normalized, Valid: ok}, true
}

type stagedEntityHost struct {
	Row    portableHostRow
	Claims []entityClaimCandidate
}

type stagedEntityIdentity struct {
	Row    portableIdentityRow
	Claims []entityClaimCandidate
}

type stagedEntityPreserved struct {
	Row   portablePreservedIdentifierRow
	Claim entityClaimCandidate
}

func hostClaims(row portableHostRow) []entityClaimCandidate {
	var claims []entityClaimCandidate
	for _, value := range []struct {
		kind string
		raw  *string
	}{{"aad_device_id", row.AADDeviceID}, {"fqdn", row.FQDN}, {"hostname", row.Hostname}} {
		if claim, ok := entityClaim(value.kind, value.raw); ok {
			claims = append(claims, claim)
		}
	}
	return claims
}

func identityClaims(row portableIdentityRow) []entityClaimCandidate {
	var claims []entityClaimCandidate
	for _, value := range []struct {
		kind string
		raw  *string
	}{{"aad_object_id", row.AADObjectID}, {"sid", row.SID}, {"upn", row.UPN}, {"email", row.Email}, {"sam_account_name", row.SAMAccountName}} {
		if claim, ok := entityClaim(value.kind, value.raw); ok {
			claims = append(claims, claim)
		}
	}
	return claims
}

// Concrete codecs stage each admitted row exactly as preparation parsed it.
// Row codecs project into the owner relation; node codecs into the
// transaction-local validation indexes.
type hostCodec struct{}

func (hostCodec) Encode(row stagedEntityHost) ([]byte, error) { return json.Marshal(row) }
func (hostCodec) Decode(line []byte) (stagedEntityHost, error) {
	var row stagedEntityHost
	return row, decodeEntityRow(line, &row)
}
func (hostCodec) Columns() []string                          { return portableHostColumns }
func (hostCodec) Values(row stagedEntityHost) ([]any, error) { return portableHostValues(row), nil }

type hostNodeCodec struct{ hostCodec }

func (hostNodeCodec) Columns() []string { return entityNodeColumns }
func (hostNodeCodec) Values(row stagedEntityHost) ([]any, error) {
	return []any{"host", row.Row.RecordID, row.Row.State}, nil
}

type identityCodec struct{}

func (identityCodec) Encode(row stagedEntityIdentity) ([]byte, error) { return json.Marshal(row) }
func (identityCodec) Decode(line []byte) (stagedEntityIdentity, error) {
	var row stagedEntityIdentity
	return row, decodeEntityRow(line, &row)
}
func (identityCodec) Columns() []string { return portableIdentityColumns }
func (identityCodec) Values(row stagedEntityIdentity) ([]any, error) {
	return portableIdentityValues(row), nil
}

type identityNodeCodec struct{ identityCodec }

func (identityNodeCodec) Columns() []string { return entityNodeColumns }
func (identityNodeCodec) Values(row stagedEntityIdentity) ([]any, error) {
	return []any{"identity", row.Row.RecordID, row.Row.State}, nil
}

type mentionCodec struct{}

func (mentionCodec) Encode(row portableMentionRow) ([]byte, error) { return json.Marshal(row) }
func (mentionCodec) Decode(line []byte) (portableMentionRow, error) {
	var row portableMentionRow
	return row, decodeEntityRow(line, &row)
}
func (mentionCodec) Columns() []string { return portableMentionColumns }
func (mentionCodec) Values(row portableMentionRow) ([]any, error) {
	return []any{row.MentionID, row.SourceRecordID, row.EntityType, row.SourceFieldKey,
		row.OriginKind, row.OriginLocator, row.RawText, row.NormalizedText,
		row.ResolutionStatus, row.RowVersion, row.Ordinal, row.RuntimeCreatedByID,
		row.CreatedAt, row.ResolvedRecordID, row.RuntimeResolvedByID, row.ResolvedAt,
		row.ResolutionMethod}, nil
}

type mentionNodeCodec struct{ mentionCodec }

func (mentionNodeCodec) Columns() []string { return []string{"id"} }
func (mentionNodeCodec) Values(row portableMentionRow) ([]any, error) {
	return []any{row.MentionID}, nil
}

type preservedCodec struct{}

func (preservedCodec) Encode(row stagedEntityPreserved) ([]byte, error) { return json.Marshal(row) }
func (preservedCodec) Decode(line []byte) (stagedEntityPreserved, error) {
	var row stagedEntityPreserved
	return row, decodeEntityRow(line, &row)
}
func (preservedCodec) Columns() []string { return portablePreservedColumns }
func (preservedCodec) Values(staged stagedEntityPreserved) ([]any, error) {
	row := staged.Row
	return []any{row.PreservedID, row.IncidentID, row.RecordID, row.EntityType, row.IdentifierType,
		row.RawValue, row.NormalizedValue, row.Classification, row.RuntimeCreatedByID,
		row.CreatedAt, row.DeletedAt}, nil
}

type aliasCodec struct{}

func (aliasCodec) Encode(row portableAliasRow) ([]byte, error) { return json.Marshal(row) }
func (aliasCodec) Decode(line []byte) (portableAliasRow, error) {
	var row portableAliasRow
	return row, decodeEntityRow(line, &row)
}
func (aliasCodec) Columns() []string { return portableAliasColumns }
func (aliasCodec) Values(row portableAliasRow) ([]any, error) {
	return []any{row.AliasID, row.IncidentID, row.RecordID, row.EntityType, row.RawText,
		row.NormalizedText, row.Classification, row.RuntimeCreatedByID,
		row.CreatedAt, row.DeletedAt}, nil
}

var entityNodeColumns = []string{"kind", "id", "state"}

func decodeEntityRow(line []byte, target any) error {
	if err := json.Unmarshal(line, target); err != nil {
		return incidentportability.ErrPreparedStream
	}
	return nil
}

func readEntityRows[T any](ctx context.Context, p preparedEntityImport, input sourceport.ImportContext, stream incidentportability.PreparedRowStream, codec incidentportability.PreparedRowCodec[T], visit func(T) error) error {
	if !p.binding.matches(input) || !p.staged || input.Preparation == nil {
		return entitySourceFailure(entitySourceIdentity)
	}
	return incidentportability.ReadPreparedRows(ctx, input.Preparation, stream, codec, visit)
}

// readEntityBatches visits a staged stream at most entitySetBatch rows at a
// time, for one set-based statement per batch.
func readEntityBatches[T any](ctx context.Context, p preparedEntityImport, input sourceport.ImportContext, stream incidentportability.PreparedRowStream, codec incidentportability.PreparedRowCodec[T], visit func([]T) error) error {
	batch := make([]T, 0, entitySetBatch)
	if err := readEntityRows(ctx, p, input, stream, codec, func(row T) error {
		batch = append(batch, row)
		if len(batch) < entitySetBatch {
			return nil
		}
		err := visit(batch)
		batch = batch[:0]
		return err
	}); err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}
	return visit(batch)
}

// stage writes the admitted rows, their identifier claims and the historical
// attribution, and binds the sealed source catalog digest.
func (p *preparedEntityImport) stage(ctx context.Context, input sourceport.ImportContext) error {
	store := input.Preparation
	if store == nil {
		return errors.New("entities: incident bundle preparation staging is required")
	}
	raw, err := json.Marshal(p.members.Members())
	if err != nil {
		return err
	}
	digest := sha256.Sum256(raw)
	p.sourceDigest = hex.EncodeToString(digest[:])
	if p.hostRows, err = incidentportability.StagePreparedRows(ctx, store, preparedHostsPath, hostCodec{}, func(emit func(stagedEntityHost) error) error {
		return p.walkHosts(ctx, input, func(row portableHostRow) error {
			return emit(stagedEntityHost{Row: row, Claims: hostClaims(row)})
		})
	}); err != nil {
		return err
	}
	if p.identityRows, err = incidentportability.StagePreparedRows(ctx, store, preparedIdentitiesPath, identityCodec{}, func(emit func(stagedEntityIdentity) error) error {
		return p.walkIdentities(ctx, input, func(row portableIdentityRow) error {
			return emit(stagedEntityIdentity{Row: row, Claims: identityClaims(row)})
		})
	}); err != nil {
		return err
	}
	if p.mentionRows, err = incidentportability.StagePreparedRows(ctx, store, preparedMentionsPath, mentionCodec{}, func(emit func(portableMentionRow) error) error {
		return p.walkMentions(ctx, input, emit)
	}); err != nil {
		return err
	}
	if p.preservedRows, err = incidentportability.StagePreparedRows(ctx, store, preparedPreservedPath, preservedCodec{}, func(emit func(stagedEntityPreserved) error) error {
		return p.walkPreserved(ctx, input, func(row portablePreservedIdentifierRow) error {
			claim, _ := entityClaim(row.IdentifierType, &row.RawValue)
			return emit(stagedEntityPreserved{Row: row, Claim: claim})
		})
	}); err != nil {
		return err
	}
	if p.aliasRows, err = incidentportability.StagePreparedRows(ctx, store, preparedAliasesPath, aliasCodec{}, func(emit func(portableAliasRow) error) error {
		return p.walkAliases(ctx, input, emit)
	}); err != nil {
		return err
	}
	admitted := func(actorID string) bool { _, ok := input.Actors.Lookup(actorID); return ok }
	p.attribution, err = incidentportability.StageAttributions(ctx, store, preparedEntityAttribute, admitted, func(emit func(incidentportability.AttributionRow) error) error {
		row := func(table, id, column string, actor *uuid.UUID) error {
			if actor == nil {
				return nil
			}
			return emit(incidentportability.AttributionRow{SourceTable: table, SourceRowID: id, SourceColumn: column, SourceActorID: actor.String()})
		}
		for _, column := range []string{"created_by_user_id", "updated_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, p.hostRows, hostCodec{}, func(host stagedEntityHost) error {
				actor := host.Row.PortableCreatedByID
				if column == "updated_by_user_id" {
					actor = host.Row.PortableUpdatedByID
				}
				return row("hosts", host.Row.RecordID.String(), column, &actor)
			}); err != nil {
				return err
			}
			if err := incidentportability.ReadPreparedRows(ctx, store, p.identityRows, identityCodec{}, func(identity stagedEntityIdentity) error {
				actor := identity.Row.PortableCreatedByID
				if column == "updated_by_user_id" {
					actor = identity.Row.PortableUpdatedByID
				}
				return row("identities", identity.Row.RecordID.String(), column, &actor)
			}); err != nil {
				return err
			}
		}
		for _, column := range []string{"created_by_user_id", "resolved_by_user_id"} {
			if err := incidentportability.ReadPreparedRows(ctx, store, p.mentionRows, mentionCodec{}, func(mention portableMentionRow) error {
				actor := &mention.PortableCreatedByID
				if column == "resolved_by_user_id" {
					actor = mention.PortableResolvedByID
				}
				return row("entity_mentions", mention.MentionID.String(), column, actor)
			}); err != nil {
				return err
			}
		}
		if err := incidentportability.ReadPreparedRows(ctx, store, p.preservedRows, preservedCodec{}, func(preserved stagedEntityPreserved) error {
			return row("entity_preserved_identifiers", preserved.Row.PreservedID.String(), "created_by_user_id", &preserved.Row.PortableCreatedByID)
		}); err != nil {
			return err
		}
		return incidentportability.ReadPreparedRows(ctx, store, p.aliasRows, aliasCodec{}, func(alias portableAliasRow) error {
			return row("entity_aliases", alias.AliasID.String(), "created_by_user_id", &alias.PortableCreatedByID)
		})
	})
	if err != nil {
		return err
	}
	p.staged = true
	return nil
}
