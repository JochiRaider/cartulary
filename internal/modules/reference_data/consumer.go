package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type PackProvenance struct {
	PackSetID             string               `json:"pack_set_id"`
	PackKey               string               `json:"pack_key"`
	PackVersion           string               `json:"pack_version"`
	ManifestSHA256        string               `json:"manifest_sha256"`
	PayloadSHA256         string               `json:"payload_sha256"`
	PackContractVersion   string               `json:"pack_contract_version"`
	ContentProfileID      string               `json:"content_profile_id"`
	ContentProfileVersion string               `json:"content_profile_version"`
	AuthorityClass        string               `json:"authority_class"`
	SourceProfileID       string               `json:"source_profile_id"`
	SourceProfileSHA256   string               `json:"source_profile_sha256"`
	SourceIdentifier      string               `json:"source_identifier"`
	SourceVersion         string               `json:"source_version"`
	SourceAsOf            *string              `json:"source_as_of"`
	SourceArtifacts       []PackSourceArtifact `json:"source_artifacts"`
	License               PackLicense          `json:"license"`
	VerificationMethod    string               `json:"verification_method"`
	LastVerifiedAt        time.Time            `json:"last_verified_at"`
	TrustValidUntil       *time.Time           `json:"trust_valid_until"`
	VerifiedSignerKeyIDs  []string             `json:"verified_signer_key_ids"`
}

type ConsumerError struct {
	Code       string  `json:"code"`
	ReasonCode *string `json:"reason_code"`
}

func (e *ConsumerError) Error() string         { return "reference pack consumer: " + e.Code }
func consumerError(code string) *ConsumerError { return &ConsumerError{Code: code} }

// ConsumerResult has exactly one populated arm. A failure never echoes input
// or invents provenance or algorithm metadata for an unresolved type.
type ConsumerResult[T any] struct {
	Value *T             `json:"value"`
	Error *ConsumerError `json:"error"`
}

func consumerSuccess[T any](value T) ConsumerResult[T] { return ConsumerResult[T]{Value: &value} }
func consumerFailure[T any](err error) ConsumerResult[T] {
	var public *ConsumerError
	if errors.As(err, &public) {
		return ConsumerResult[T]{Error: public}
	}
	var format *packformat.Failure
	if errors.As(err, &format) {
		code := format.Code
		if code == "indicator_input_too_long" {
			return ConsumerResult[T]{Error: &ConsumerError{Code: "invalid_pack_request", ReasonCode: &code}}
		}
		return ConsumerResult[T]{Error: consumerError(code)}
	}
	return ConsumerResult[T]{Error: consumerError("pack_unavailable")}
}

type GetPackEntryRequest struct {
	PackSetID string `json:"pack_set_id"`
	PackKey   string `json:"pack_key"`
	EntryID   string `json:"entry_id"`
}
type LookupPackEntriesRequest struct {
	PackSetID   string  `json:"pack_set_id"`
	PackKey     string  `json:"pack_key"`
	LookupKind  string  `json:"lookup_kind"`
	LookupValue any     `json:"lookup_value"`
	Limit       *int    `json:"limit,omitempty"`
	Cursor      *string `json:"cursor,omitempty"`
}
type EvaluateIndicatorRequest struct {
	PackSetID       string `json:"pack_set_id"`
	IndicatorTypeID string `json:"indicator_type_id"`
	ValueKind       string `json:"value_kind"`
	RawValue        string `json:"raw_value"`
}
type GetPackProvenanceRequest struct {
	PackSetID string `json:"pack_set_id"`
	PackKey   string `json:"pack_key"`
}
type PackEntry struct {
	Item       json.RawMessage `json:"item"`
	Provenance PackProvenance  `json:"provenance"`
}
type PackEntryPage struct {
	Items      []json.RawMessage `json:"items"`
	NextCursor *string           `json:"next_cursor"`
	Provenance PackProvenance    `json:"provenance"`
}
type IndicatorEvaluation struct {
	Evaluation
	Provenance PackProvenance `json:"provenance"`
}

type Consumer interface {
	ResolveCurrentPackSet(context.Context) ConsumerResult[PackSet]
	GetPackEntry(context.Context, GetPackEntryRequest) ConsumerResult[PackEntry]
	LookupPackEntries(context.Context, LookupPackEntriesRequest) ConsumerResult[PackEntryPage]
	EvaluateIndicatorValue(context.Context, EvaluateIndicatorRequest) ConsumerResult[IndicatorEvaluation]
	GetPackProvenance(context.Context, GetPackProvenanceRequest) ConsumerResult[PackProvenance]
}

type indexedItem struct {
	Item     json.RawMessage
	Position string
}
type consumerRepository interface {
	CurrentSet(context.Context) (PackSet, error)
	RetainedSet(context.Context, string) (PackSet, error)
	Provenance(context.Context, string, string) (PackProvenance, error)
	Entry(context.Context, PackSetMember, string) (json.RawMessage, error)
	Lookup(context.Context, PackSetMember, packformat.LookupInput, string, int) ([]indexedItem, error)
}
type packConsumer struct {
	repository consumerRepository
	codec      *pagination.Codec
	now        func() time.Time
}

var setIDPattern = regexp.MustCompile(`^rpset_[0-9a-f]{64}$`)
var consumerPackKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$`)

func newPackConsumer(repository consumerRepository, codec *pagination.Codec, now func() time.Time) Consumer {
	return &packConsumer{repository: repository, codec: codec, now: now}
}
func (c *packConsumer) ResolveCurrentPackSet(ctx context.Context) ConsumerResult[PackSet] {
	set, err := c.repository.CurrentSet(ctx)
	if err != nil {
		return consumerFailure[PackSet](err)
	}
	return consumerSuccess(set)
}
func validConsumerKey(key string) bool {
	return len(key) <= 128 && consumerPackKeyPattern.MatchString(key)
}
func (c *packConsumer) resolve(ctx context.Context, setID, key string) (PackSetMember, PackProvenance, error) {
	set, err := c.repository.RetainedSet(ctx, setID)
	if err != nil {
		return PackSetMember{}, PackProvenance{}, err
	}
	for _, member := range set.Members {
		if member.Key == key {
			provenance, err := c.repository.Provenance(ctx, setID, key)
			return member, provenance, err
		}
	}
	return PackSetMember{}, PackProvenance{}, consumerError("pack_unavailable")
}
func (c *packConsumer) GetPackProvenance(ctx context.Context, request GetPackProvenanceRequest) ConsumerResult[PackProvenance] {
	if err := packformat.ValidateConsumerRequest("GetPackProvenance", map[string]any{"pack_set_id": request.PackSetID, "pack_key": request.PackKey}); err != nil {
		return consumerFailure[PackProvenance](err)
	}
	_, provenance, err := c.resolve(ctx, request.PackSetID, request.PackKey)
	if err != nil {
		return consumerFailure[PackProvenance](err)
	}
	return consumerSuccess(provenance)
}
func (c *packConsumer) GetPackEntry(ctx context.Context, request GetPackEntryRequest) ConsumerResult[PackEntry] {
	if err := packformat.ValidateConsumerRequest("GetPackEntry", map[string]any{"pack_set_id": request.PackSetID, "pack_key": request.PackKey, "entry_id": request.EntryID}); err != nil {
		return consumerFailure[PackEntry](err)
	}
	member, provenance, err := c.resolve(ctx, request.PackSetID, request.PackKey)
	if err != nil {
		return consumerFailure[PackEntry](err)
	}
	if err := packformat.ValidateEntryIdentity(member.Key, request.EntryID); err != nil {
		return consumerFailure[PackEntry](err)
	}
	item, err := c.repository.Entry(ctx, member, request.EntryID)
	if err != nil {
		return consumerFailure[PackEntry](err)
	}
	return consumerSuccess(PackEntry{Item: item, Provenance: provenance})
}
func (c *packConsumer) EvaluateIndicatorValue(ctx context.Context, request EvaluateIndicatorRequest) ConsumerResult[IndicatorEvaluation] {
	if err := packformat.ValidateConsumerRequest("EvaluateIndicatorValue", map[string]any{"pack_set_id": request.PackSetID, "indicator_type_id": request.IndicatorTypeID, "value_kind": request.ValueKind, "raw_value": request.RawValue}); err != nil {
		return consumerFailure[IndicatorEvaluation](err)
	}
	member, provenance, err := c.resolve(ctx, request.PackSetID, "type_registry.indicator")
	if err != nil {
		return consumerFailure[IndicatorEvaluation](err)
	}
	row, err := c.repository.Entry(ctx, member, request.IndicatorTypeID)
	if err != nil {
		var public *ConsumerError
		if errors.As(err, &public) && public.Code == "entry_not_found" {
			return consumerFailure[IndicatorEvaluation](consumerError("indicator_type_unsupported"))
		}
		return consumerFailure[IndicatorEvaluation](err)
	}
	var entry struct {
		ID        string   `json:"entry_id"`
		Kinds     []string `json:"allowed_value_kinds"`
		Normalize string   `json:"normalization_algorithm_id"`
		Validate  string   `json:"validation_algorithm_id"`
		Defang    string   `json:"defang_algorithm_id"`
		Dedupe    string   `json:"dedupe_algorithm_id"`
	}
	if err := json.Unmarshal(row, &entry); err != nil {
		return consumerFailure[IndicatorEvaluation](err)
	}
	evaluation, err := packformat.Evaluate(packformat.IndicatorPolicy{Type: entry.ID, Kinds: entry.Kinds, Normalize: entry.Normalize, Validate: entry.Validate, Defang: entry.Defang, Dedupe: entry.Dedupe}, request.ValueKind, request.RawValue)
	if err != nil {
		return consumerFailure[IndicatorEvaluation](err)
	}
	return consumerSuccess(IndicatorEvaluation{Evaluation: Evaluation(evaluation), Provenance: provenance})
}
func (c *packConsumer) LookupPackEntries(ctx context.Context, request LookupPackEntriesRequest) ConsumerResult[PackEntryPage] {
	now := c.now().UTC()
	bad := func(code string) ConsumerResult[PackEntryPage] {
		return consumerFailure[PackEntryPage](consumerError(code))
	}
	fields := map[string]any{"pack_set_id": request.PackSetID, "pack_key": request.PackKey, "lookup_kind": request.LookupKind, "lookup_value": request.LookupValue}
	if request.Limit != nil {
		fields["limit"] = json.Number(strconv.Itoa(*request.Limit))
	}
	if request.Cursor != nil {
		fields["cursor"] = *request.Cursor
	}
	if err := packformat.ValidateConsumerRequest("LookupPackEntries", fields); err != nil {
		return consumerFailure[PackEntryPage](err)
	}
	member, provenance, err := c.resolve(ctx, request.PackSetID, request.PackKey)
	if err != nil {
		return consumerFailure[PackEntryPage](err)
	}
	query, err := packformat.NormalizeLookup(member.Key, request.LookupKind, request.LookupValue)
	if err != nil {
		return consumerFailure[PackEntryPage](err)
	}
	limit := 50
	if request.Limit != nil {
		limit = *request.Limit
	}
	scope := map[string]string{"set": request.PackSetID, "pack": request.PackKey, "kind": query.Kind, "value": query.Value}
	position := ""
	if request.Cursor != nil {
		cursor, err := c.codec.Decode(*request.Cursor)
		if err != nil || cursor.Route != "reference_pack.lookup.v1" || cursor.Mode != pagination.ModeKeyset || cursor.ActorUserID != "reference_pack_consumer" || len(cursor.Scope) != 4 || len(cursor.Position) != 3 || cursor.Limit < 1 || cursor.Limit > 200 {
			return bad("cursor_invalid")
		}
		issued, e1 := time.Parse(time.RFC3339Nano, cursor.Position["issued"])
		expires, e2 := time.Parse(time.RFC3339Nano, cursor.Position["expires"])
		if e1 != nil || e2 != nil || issued.Format(time.RFC3339Nano) != cursor.Position["issued"] || expires.Format(time.RFC3339Nano) != cursor.Position["expires"] || !expires.Equal(issued.Add(900*time.Second)) || !now.Before(expires) || now.Before(issued) {
			return bad("cursor_invalid")
		}
		for key, value := range scope {
			if cursor.Scope[key] != value {
				return bad("cursor_query_mismatch")
			}
		}
		if request.Limit != nil && limit != cursor.Limit {
			return bad("cursor_query_mismatch")
		}
		limit = cursor.Limit
		position = cursor.Position["after"]
		if position == "" {
			return bad("cursor_invalid")
		}
	}
	rows, err := c.repository.Lookup(ctx, member, query, position, limit+1)
	if err != nil {
		return consumerFailure[PackEntryPage](err)
	}
	page := PackEntryPage{Items: []json.RawMessage{}, Provenance: provenance}
	if len(rows) > limit {
		rows = rows[:limit]
		token, err := c.codec.Encode(pagination.Cursor{Mode: pagination.ModeKeyset, Route: "reference_pack.lookup.v1", ActorUserID: "reference_pack_consumer", Limit: limit, Scope: scope, Position: map[string]string{"after": rows[len(rows)-1].Position, "issued": now.Format(time.RFC3339Nano), "expires": now.Add(900 * time.Second).Format(time.RFC3339Nano)}})
		if err != nil {
			return consumerFailure[PackEntryPage](err)
		}
		page.NextCursor = &token
	}
	for _, row := range rows {
		page.Items = append(page.Items, row.Item)
	}
	return consumerSuccess(page)
}

// Decode the original bytes before ordinary struct decoding can erase duplicate
// keys, null/omission distinctions or malformed Unicode. Decode into a temporary
// value so rejected input never partly mutates a reused request.
func decodeConsumerRequest(operation string, data []byte, target any) error {
	if err := packformat.AdmitConsumerRequest(operation, data); err != nil {
		return consumerFailure[any](err).Error
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return consumerError("invalid_pack_request")
	}
	return nil
}

func (r *GetPackEntryRequest) UnmarshalJSON(data []byte) error {
	type admitted GetPackEntryRequest
	var value admitted
	if err := decodeConsumerRequest("GetPackEntry", data, &value); err != nil {
		return err
	}
	*r = GetPackEntryRequest(value)
	return nil
}
func (r *LookupPackEntriesRequest) UnmarshalJSON(data []byte) error {
	type admitted LookupPackEntriesRequest
	var value admitted
	if err := decodeConsumerRequest("LookupPackEntries", data, &value); err != nil {
		return err
	}
	*r = LookupPackEntriesRequest(value)
	return nil
}
func (r *EvaluateIndicatorRequest) UnmarshalJSON(data []byte) error {
	type admitted EvaluateIndicatorRequest
	var value admitted
	if err := decodeConsumerRequest("EvaluateIndicatorValue", data, &value); err != nil {
		return err
	}
	*r = EvaluateIndicatorRequest(value)
	return nil
}
func (r *GetPackProvenanceRequest) UnmarshalJSON(data []byte) error {
	type admitted GetPackProvenanceRequest
	var value admitted
	if err := decodeConsumerRequest("GetPackProvenance", data, &value); err != nil {
		return err
	}
	*r = GetPackProvenanceRequest(value)
	return nil
}
