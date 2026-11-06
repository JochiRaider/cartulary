package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

var errPortableNetworkFlow = errors.New("network flow portable state is invalid")

type preparedNetworkFlow struct {
	incidentID  uuid.UUID
	operationID string
	actorID     uuid.UUID
	actors      sourceport.ActorCatalog
	members     *incidentportability.BoundMembers
	tables      map[string]tableRecord
	counts      map[string]int64
	limits      EffectiveLimits
	graphs      []preparedPortableGraph
}

type preparedPortableGraph struct {
	declaration graphViewDeclaration
	members     *incidentportability.BoundMembers
	path        string
}

func walkPortable[T any](ctx context.Context, members incidentportability.MemberReader, path, required, nullable string, visit func(T) error) error {
	return incidentportability.WalkPath(ctx, members, portabilityMemberPrefix+path, func(object map[string]any) error {
		raw, err := json.Marshal(object)
		if err != nil {
			return err
		}
		var value T
		if _, err := decodeClosedImportObject(raw, required, "", nullable, &value); err != nil {
			return fmt.Errorf("%s record shape: %w", path, errPortableNetworkFlow)
		}
		// A typed round trip must preserve every nested member. Unknown nested
		// fields and omitted required struct fields cannot disappear on import.
		closed, err := json.Marshal(value)
		if err != nil || !canonicalEqual(json.RawMessage(raw), json.RawMessage(closed)) {
			return fmt.Errorf("%s nested record shape: %w", path, errPortableNetworkFlow)
		}
		if err := visit(value); err != nil {
			return fmt.Errorf("%s record admission: %w", path, err)
		}
		return nil
	})
}
func validPortableTime(value time.Time) bool {
	return value.Year() >= 1 && value.Year() <= 9999 && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
func (p *preparedNetworkFlow) actor(id uuid.UUID) bool {
	_, ok := p.actors.Lookup(id.String())
	return id != uuid.Nil && ok
}

func prepareNetworkFlow(ctx context.Context, invocation incidentbundles.ImportInvocation, limits EffectiveLimits) (*preparedNetworkFlow, error) {
	if invocation.ProfileID != ProfileID || invocation.ContractMajor != 10 || invocation.StateVersion != 7 || invocation.ContentSchemaID != portabilityContentSchema || invocation.ClaimState != incidentbundles.ClaimStateClaimed || invocation.IncidentID == uuid.Nil || invocation.ActorUserID == uuid.Nil || invocation.OperationID == "" {
		return nil, errPortableNetworkFlow
	}
	dependencies, ok := invocation.Dependencies.(interface {
		CommittedSources() committedPortabilitySources
		IndicatorReferences() portableIndicatorReferences
	})
	if !ok || invocation.Dependencies.IncidentScope() != invocation.IncidentID {
		return nil, incidentbundles.ErrPortabilityUnavailable
	}
	paths := make([]string, 0, len(portabilityFamilyPaths))
	for _, path := range portabilityFamilyPaths {
		paths = append(paths, portabilityMemberPrefix+path)
	}
	if invocation.Members == nil || len(invocation.Members.Members()) != len(paths) {
		return nil, errPortableNetworkFlow
	}
	members, err := incidentportability.BindMembers(invocation.Members, paths)
	if err != nil {
		return nil, err
	}
	p := &preparedNetworkFlow{incidentID: invocation.IncidentID, operationID: invocation.OperationID, actorID: invocation.ActorUserID, actors: invocation.Actors, members: members, tables: map[string]tableRecord{}, counts: map[string]int64{"tables": 0, "observations": 0, "diagnostics": 0, "indicator_bindings": 0, "graph_views": 0}, limits: limits}
	previous := ""
	var active int64
	sourceIDs := map[uuid.UUID]bool{}
	names := map[string]bool{}
	err = walkPortable(ctx, members, "tables.ndjson", portableTableMembers, "deleted_at", func(v portableTable) error {
		if v.TableID <= previous || int64(len(p.tables)) >= limits.MaxRetainedTablesPerIncident || !p.actor(v.CreatedByUserID) || v.IncidentID != p.incidentID {
			return errPortableNetworkFlow
		}
		previous = v.TableID
		table := v.retained()
		name, nameErr := normalizeExplicitDisplayName(table.DisplayName)
		profile, known := mappingSourceProfile(table.SourceProfileID)
		if nameErr != nil || name != table.DisplayName || !known || profile.ParserProfileID != table.ParserProfileID || !linkTableIDPattern.MatchString(table.TableID) || table.TableVersion < 1 || table.CommittedSourceID == uuid.Nil || sourceIDs[table.CommittedSourceID] || !validLowerSHA256(table.SourceInterpretationSHA256) || !validLowerSHA256(table.SourceContentSHA256) || !validLowerSHA256(table.MappingFingerprint) || table.RowCountAccepted < 1 || table.RowCountAccepted > limits.MaxAcceptedRowsPerTable || table.RowCountRejected < 0 || table.RowCountRejected > limits.MaxRowsPerCSV-table.RowCountAccepted || table.SourceRawCellScalarLimit < 1 || table.SourceRawCellScalarLimit > 16384 || utf8.RuneCountInString(table.SourceFilenameDisplay) > 256 || !validPortableTime(table.CreatedAt) || !validPortableTime(table.UpdatedAt) || table.UpdatedAt.Before(table.CreatedAt) {
			return errPortableNetworkFlow
		}
		switch table.TableStatus {
		case tableStatusActive:
			active++
			if table.DeletedAt != nil || active > limits.MaxActiveTablesPerIncident || names[table.DisplayName] {
				return errPortableNetworkFlow
			}
			names[table.DisplayName] = true
		case "soft_deleted":
			if table.DeletedAt == nil || !validPortableTime(*table.DeletedAt) || table.DeletedAt.Before(table.CreatedAt) || table.DeletedAt.After(table.UpdatedAt) {
				return errPortableNetworkFlow
			}
		default:
			return errPortableNetworkFlow
		}
		source, err := dependencies.CommittedSources().ReadSource(ctx, table.CommittedSourceID)
		if err != nil {
			return err
		}
		if source.IncidentID != p.incidentID || source.ContentSHA256 != table.SourceContentSHA256 || source.MappingFingerprint != table.MappingFingerprint || source.TargetKind != targetKindNetworkFlowTable || source.ExtensionProfileID != ProfileID || source.SourceFileKind != "csv" {
			return errPortableNetworkFlow
		}
		var envelope struct {
			OwnerMapping         json.RawMessage `json:"owner_mapping"`
			OwnerMappingSchemaID string          `json:"owner_mapping_schema_id"`
		}
		if json.Unmarshal(source.ApprovedMapping, &envelope) != nil || envelope.OwnerMappingSchemaID != importMappingSchemaID {
			return errPortableNetworkFlow
		}
		interpretationLimits := limits
		interpretationLimits.MaxRawCellScalarLength = table.SourceRawCellScalarLimit
		mapping, err := decodeApprovedMapping(envelope.OwnerMapping, interpretationLimits)
		if err != nil {
			return err
		}
		fingerprint, err := mappingFingerprint(mapping, source.ContentSHA256)
		if err != nil || fingerprint != table.MappingFingerprint || mapping.SourceProfileID != table.SourceProfileID || mapping.ParserProfileID != table.ParserProfileID {
			return errPortableNetworkFlow
		}
		sourceIDs[table.CommittedSourceID] = true
		p.tables[table.TableID] = table
		p.counts["tables"]++
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(p.tables) == 0 {
		return nil, errPortableNetworkFlow
	}
	if err := p.verifyCommittedObservations(ctx, dependencies.CommittedSources()); err != nil {
		return nil, err
	}
	if err := p.validateBindings(ctx, dependencies.IndicatorReferences()); err != nil {
		return nil, err
	}
	if err := p.prepareGraphs(ctx, invocation.PreparationMembers); err != nil {
		return nil, err
	}
	return p, nil
}

func comparePortableRows(a, b flowRow) int {
	if value := strings.Compare(a.NetworkFlowTableID, b.NetworkFlowTableID); value != 0 {
		return value
	}
	if a.SourceRowNumber < b.SourceRowNumber {
		return -1
	}
	if a.SourceRowNumber > b.SourceRowNumber {
		return 1
	}
	return 0
}

func (p *preparedNetworkFlow) validateObservation(row flowRow, table tableRecord) error {
	if row.IncidentID != p.incidentID || row.MappingFingerprint != table.MappingFingerprint || !p.actor(row.CreatedByUserID) || !validPortableTime(row.CreatedAt) || !row.CreatedAt.Equal(table.CreatedAt) || !validLowerSHA256(row.SourceRowDigestSHA256) || !validLowerSHA256(row.NormalizedRowDigestSHA256) || row.RowID != rowID(row.IncidentID, row.NetworkFlowTableID, row.SourceRowNumber, row.SourceRowDigestSHA256, row.NormalizedRowDigestSHA256) || !validPortableTime(row.FlowStartUTC) || !validPortableTime(row.FlowEndUTC) || row.FlowEndUTC.Before(row.FlowStartUTC) || row.IPProtocol < 0 || row.IPProtocol > 255 {
		return fmt.Errorf("observation metadata: %w", errPortableNetworkFlow)
	}
	if err := validateObservationDirection(flowEndpoint{row.EndpointAIP, row.EndpointAPort}, flowEndpoint{row.EndpointBIP, row.EndpointBPort}, row.Direction, row.DirectionEvidence, table.SourceProfileID); err != nil {
		return err
	}
	for _, port := range []*int32{row.EndpointAPort, row.EndpointBPort} {
		if port != nil && (*port < 0 || *port > 65535) {
			return errPortableNetworkFlow
		}
	}
	for _, value := range []string{row.BytesCount, row.PacketsCount} {
		if !unsignedDecimalRE.MatchString(value) || len(value) > 20 || (len(value) == 20 && value > "18446744073709551615") {
			return errPortableNetworkFlow
		}
	}
	var retained retainedRawValues
	if json.Unmarshal(row.UnmappedRaw, &retained) != nil || retained == nil {
		return errPortableNetworkFlow
	}
	values := rowResource(row)
	values[fieldDirectionEvidence] = row.DirectionEvidence
	values[fieldEndpointAIP] = row.EndpointAIP
	values[fieldEndpointBIP] = row.EndpointBIP
	values[fieldEndpointAPort] = nullableInt32Value(row.EndpointAPort)
	values[fieldEndpointBPort] = nullableInt32Value(row.EndpointBPort)
	values[fieldObservationSourceRef] = nil
	digest, err := normalizedRowDigest(row.MappingFingerprint, values, retained)
	if err != nil || digest != row.NormalizedRowDigestSHA256 {
		return fmt.Errorf("normalized observation digest: %w", errPortableNetworkFlow)
	}
	want := map[string]any{"committed_source_id": table.CommittedSourceID.String(), "source_content_sha256": table.SourceContentSHA256, "source_profile_id": table.SourceProfileID, "parser_profile_id": table.ParserProfileID, "mapping_fingerprint": table.MappingFingerprint, "source_row_number": row.SourceRowNumber, "source_row_digest_sha256": row.SourceRowDigestSHA256}
	if !canonicalEqual(row.ObservationSourceRef, want) {
		return fmt.Errorf("observation source reference: %w", errPortableNetworkFlow)
	}
	return nil
}

func (p *preparedNetworkFlow) validateDiagnostics(ctx context.Context) error {
	previous := ""
	counts := map[string]int64{}
	rejected := map[string]int64{}
	current := ""
	var ordinals []byte
	err := walkPortable(ctx, p.members, "diagnostics.ndjson", portableDiagnosticMembers, "", func(v portableDiagnostic) error {
		table, ok := p.tables[v.TableID]
		key := v.TableID + "\x00" + v.Diagnostic.DiagnosticID
		if !ok || v.IncidentID != p.incidentID || key <= previous || !validPortableTime(v.CreatedAt) || !v.CreatedAt.Equal(table.CreatedAt) || validateDiagnostic(v.Diagnostic, table.SourceContentSHA256, table.MappingFingerprint) != nil {
			return errPortableNetworkFlow
		}
		if current != v.TableID {
			current = v.TableID
			ordinals = make([]byte, (p.limits.MaxRowsPerCSV+9)/8)
		}
		n := v.Diagnostic.SourceRowNumber
		if n < 2 || n > p.limits.MaxRowsPerCSV+1 {
			return errPortableNetworkFlow
		}
		i, mask := n/8, byte(1<<uint(n%8))
		if ordinals[i]&mask == 0 {
			rejected[v.TableID]++
			ordinals[i] |= mask
		}
		previous = key
		counts[v.TableID]++
		if counts[v.TableID] > p.limits.MaxRejectedRowDiagnostics {
			return incidentbundles.ErrPortabilityLimit
		}
		p.counts["diagnostics"]++
		return nil
	})
	if err != nil {
		return err
	}
	for id, t := range p.tables {
		if rejected[id] > t.RowCountRejected || (!t.DiagnosticsTruncated && rejected[id] != t.RowCountRejected) {
			return errPortableNetworkFlow
		}
	}
	return nil
}
func (p *preparedNetworkFlow) validateBindings(ctx context.Context, indicators portableIndicatorReferences) error {
	previous := ""
	return walkPortable(ctx, p.members, "indicator_bindings.ndjson", portableBindingMembers, "", func(v portableBinding) error {
		if v.IncidentID != p.incidentID || v.BindingID <= previous || !linkBindingIDPattern.MatchString(v.BindingID) || !p.actor(v.CreatedByUserID) || !validPortableTime(v.CreatedAt) || !validBindingSelectorKind(v.SelectorKind) || len(v.SourceRowRefs) == 0 || int64(len(v.SourceRowRefs)) > p.limits.MaxBindingSourceRowRefs || v.SourceRowRefsTotalCount < int64(len(v.SourceRowRefs)) || v.SourceRowRefsTruncated != (v.SourceRowRefsTotalCount > int64(len(v.SourceRowRefs))) {
			return errPortableNetworkFlow
		}
		if v.SelectorKind == "row_field_value" && (len(v.SourceRowRefs) != 1 || v.SourceRowRefsTruncated) || v.SelectorKind == "row_refs" && v.SourceRowRefsTruncated {
			return errPortableNetworkFlow
		}
		target, err := indicators.FindIndicator(ctx, v.IndicatorID)
		if err != nil {
			return err
		}
		if target.IncidentID != p.incidentID || validateIndicatorTargetLogical(target, v.CandidateValue, v.IndicatorType) != nil {
			return errPortableNetworkFlow
		}
		wanted := map[string]networkFlowRowRef{}
		for _, ref := range v.SourceRowRefs {
			if !validLinkRowRef(ref) {
				return errPortableNetworkFlow
			}
			if _, exists := wanted[ref.NetworkFlowRowID]; exists {
				return errPortableNetworkFlow
			}
			wanted[ref.NetworkFlowRowID] = ref
		}
		err = walkPortable(ctx, p.members, "observations.ndjson", portableObservationMembers, portableObservationNullable, func(row portableObservation) error {
			ref, exists := wanted[row.RowID]
			if !exists {
				return nil
			}
			if row.NetworkFlowTableID != ref.NetworkFlowTableID || row.SourceRowNumber != ref.SourceRowNumber || row.MappingFingerprint != ref.MappingFingerprint || (row.EndpointAIP != v.CandidateValue && row.EndpointBIP != v.CandidateValue) {
				return errPortableNetworkFlow
			}
			delete(wanted, row.RowID)
			return nil
		})
		if err != nil {
			return err
		}
		if len(wanted) != 0 {
			return errPortableNetworkFlow
		}
		previous = v.BindingID
		p.counts["indicator_bindings"]++
		return nil
	})
}
