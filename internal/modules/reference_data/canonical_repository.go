package reference_data

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// canonicalRepository is private owner persistence behind consumer and
// application operations. Other source owners receive Consumer or pin ports.
type canonicalRepository struct {
	pool       retainedContentQuery
	storage    ArtifactStorage
	invalidate func(context.Context, string, string) error
}

func NewConsumer(pool *pgxpool.Pool, storage ArtifactStorage, codec *pagination.Codec, now func() time.Time, integrity IntegrityOptions) (Consumer, error) {
	if pool == nil || storage == nil || codec == nil || now == nil {
		return nil, errors.New("reference pack: incomplete consumer dependencies")
	}
	coordinator, err := newIntegrityCoordinator(pool, storage, now, integrity)
	if err != nil {
		return nil, err
	}
	return observeConsumer(newPackConsumer(&canonicalRepository{pool: pool, storage: storage, invalidate: coordinator.InvalidateUnavailablePack}, codec, now), integrity.Observer), nil
}

func versionObjectID(key, version string) string {
	return "rpver_" + packformat.Digest([]byte(key+"\x00"+version))
}

func (r *canonicalRepository) CurrentSet(ctx context.Context) (PackSet, error) {
	var data []byte
	err := r.pool.QueryRow(ctx, `SELECT s.canonical_set FROM reference_pack_current_set c JOIN reference_pack_sets s USING(pack_set_id) WHERE c.singleton`).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return PackSet{}, consumerError("pack_unavailable")
	}
	if err != nil {
		return PackSet{}, err
	}
	return decodeRetainedSet(data)
}
func (r *canonicalRepository) RetainedSet(ctx context.Context, id string) (PackSet, error) {
	var data []byte
	err := r.pool.QueryRow(ctx, `SELECT canonical_set FROM reference_pack_sets WHERE pack_set_id=$1`, id).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return PackSet{}, consumerError("pack_set_not_found")
	}
	if err != nil {
		return PackSet{}, err
	}
	return decodeRetainedSet(data)
}
func decodeRetainedSet(data []byte) (PackSet, error) {
	var set PackSet
	if err := json.Unmarshal(data, &set); err != nil {
		return PackSet{}, err
	}
	expected, err := packformat.BuildSet(set.Members)
	if err != nil || expected.ID != set.ID || expected.SHA256 != set.SHA256 || expected.SchemaID != set.SchemaID || !slices.Equal(expected.Members, set.Members) {
		return PackSet{}, consumerError("pack_unavailable")
	}
	canonical, err := canonicaljson.Marshal(set)
	if err != nil || !bytes.Equal(canonical, data) {
		return PackSet{}, consumerError("pack_unavailable")
	}
	return set, nil
}
func (r *canonicalRepository) Provenance(ctx context.Context, setID, key string) (PackProvenance, error) {
	var data, setData []byte
	err := r.pool.QueryRow(ctx, `SELECT canonical_set,canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, setID).Scan(&setData, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		return PackProvenance{}, consumerError("pack_set_not_found")
	}
	if err != nil {
		return PackProvenance{}, err
	}
	set, err := decodeRetainedSet(setData)
	if err != nil || set.ID != setID {
		return PackProvenance{}, consumerError("pack_unavailable")
	}
	anchors, err := decodeRetainedProvenance(data, set)
	if err != nil {
		return PackProvenance{}, err
	}
	for _, anchor := range anchors {
		if anchor.PackKey != key {
			continue
		}
		if anchor.PackSetID != setID {
			return PackProvenance{}, consumerError("pack_unavailable")
		}
		if err := r.requireAvailable(ctx, setID, anchor); err != nil {
			return PackProvenance{}, err
		}
		return anchor, nil
	}
	return PackProvenance{}, consumerError("pack_unavailable")
}
func (r *canonicalRepository) requireAvailable(ctx context.Context, setID string, p PackProvenance) error {
	var health string
	var failure *string
	var pinned, indexed bool
	err := r.pool.QueryRow(ctx, `SELECT c.health,c.last_failure_code,EXISTS(SELECT 1 FROM reference_pack_pins WHERE pack_set_id=$1),
 EXISTS(SELECT 1 FROM reference_pack_index_generations i WHERE i.index_id=c.current_index_id AND i.complete AND i.manifest_sha256=v.manifest_sha256 AND i.payload_sha256=v.payload_sha256)
 FROM reference_pack_candidates c JOIN reference_pack_versions v USING(pack_key,pack_version)
 WHERE c.pack_key=$2 AND c.pack_version=$3 AND v.manifest_sha256=$4 AND v.payload_sha256=$5 AND NOT c.removed`, setID, p.PackKey, p.PackVersion, p.ManifestSHA256, p.PayloadSHA256).Scan(&health, &failure, &pinned, &indexed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return consumerError("pack_unavailable")
		}
		return err
	}
	if !indexed || health != "verified_available" && !(health == "failed" && pinned && failure != nil && *failure == "metadata_expired") {
		return consumerError("pack_unavailable")
	}
	if err := checkRetainedVersion(ctx, r.pool, r.storage, p.PackKey, p.PackVersion); err != nil {
		var rejected *ContentRejection
		if errors.As(err, &rejected) && r.invalidate != nil {
			if err := r.invalidate(ctx, p.PackKey, p.PackVersion); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				// Definitive loss remains unavailable even when its independent
				// consequence transaction cannot publish (for example, Base loss
				// has no healthy fallback). Preserve the internal cause as well.
				return errors.Join(consumerError("pack_unavailable"), err)
			}
		}
		return consumerError("pack_unavailable")
	}

	return nil
}
func (r *canonicalRepository) Entry(ctx context.Context, member PackSetMember, id string) (json.RawMessage, error) {
	var data []byte
	err := r.pool.QueryRow(ctx, `SELECT i.canonical_item FROM reference_pack_candidates c JOIN reference_pack_indexes i ON i.index_id=c.current_index_id WHERE c.pack_key=$1 AND c.pack_version=$2 AND i.entry_id=$3 AND i.entry_kind IN ('entry','object')`, member.Key, member.Version, id).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, consumerError("entry_not_found")
	}
	return json.RawMessage(data), err
}

type lookupPosition struct {
	Order []byte `json:"order"`
	ID    string `json:"id"`
}

func encodeLookupPosition(order []byte, id string) string {
	data, _ := json.Marshal(lookupPosition{order, id})
	return base64.RawURLEncoding.EncodeToString(data)
}
func decodeLookupPosition(raw string) (lookupPosition, error) {
	if raw == "" {
		return lookupPosition{}, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return lookupPosition{}, consumerError("cursor_invalid")
	}
	var p lookupPosition
	if json.Unmarshal(data, &p) != nil || p.ID == "" || len(p.ID) > 512 || len(p.Order) > 65536 {
		return lookupPosition{}, consumerError("cursor_invalid")
	}
	return p, nil
}
func (r *canonicalRepository) Lookup(ctx context.Context, member PackSetMember, q packformat.LookupInput, after string, limit int) ([]indexedItem, error) {
	p, err := decodeLookupPosition(after)
	if err != nil {
		return nil, err
	}
	var rows pgx.Rows
	if q.NetworkAddress == nil {
		digest := sha256.Sum256([]byte(q.Value))
		rows, err = r.pool.Query(ctx, `SELECT i.canonical_item,k.sort_key,k.entry_id FROM reference_pack_candidates c JOIN reference_pack_lookup_keys k ON k.index_id=c.current_index_id JOIN reference_pack_indexes i USING(index_id,entry_kind,entry_id)
 WHERE c.pack_key=$1 AND c.pack_version=$2 AND k.lookup_kind=$3 AND k.lookup_sha256=$4 AND k.lookup_value=$5 AND ($6='' OR (k.sort_key,k.entry_id COLLATE "C")>($7,$6 COLLATE "C"))
 ORDER BY k.sort_key,k.entry_id COLLATE "C" LIMIT $8`, member.Key, member.Version, q.Kind, digest[:], []byte(q.Value), p.ID, p.Order, limit)
	} else {
		// Network membership is evaluated against the validated canonical CIDR
		// identifier. Ordering remains the profile's address/prefix index order.
		rows, err = r.pool.Query(ctx, `SELECT i.canonical_item,k.sort_key,k.entry_id FROM reference_pack_candidates c JOIN reference_pack_lookup_keys k ON k.index_id=c.current_index_id JOIN reference_pack_indexes i USING(index_id,entry_kind,entry_id)
 WHERE c.pack_key=$1 AND c.pack_version=$2 AND k.lookup_kind='network' AND $3::inet <<= k.entry_id::cidr AND ($4='' OR (k.sort_key,k.entry_id COLLATE "C")>($5,$4 COLLATE "C"))
 ORDER BY k.sort_key,k.entry_id COLLATE "C" LIMIT $6`, member.Key, member.Version, q.NetworkAddress.String(), p.ID, p.Order, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []indexedItem{}
	for rows.Next() {
		var data, order []byte
		var id string
		if err := rows.Scan(&data, &order, &id); err != nil {
			return nil, err
		}
		items = append(items, indexedItem{Item: data, Position: encodeLookupPosition(order, id)})
	}
	return items, rows.Err()
}

// PinCurrentTx is called inside the consuming owner's admission transaction.
// It captures one current set and never silently rebases after locking.
func PinCurrentTx(ctx context.Context, tx pgx.Tx, ownerKind, ownerID string, operationID uuid.UUID) (PackSet, error) {
	var before string
	var data []byte
	if err := tx.QueryRow(ctx, `SELECT s.pack_set_id,s.canonical_set FROM reference_pack_current_set c JOIN reference_pack_sets s USING(pack_set_id) WHERE singleton`).Scan(&before, &data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PackSet{}, consumerError("pack_unavailable")
		}
		return PackSet{}, err
	}
	set, err := decodeRetainedSet(data)
	if err != nil {
		return PackSet{}, err
	}
	for _, member := range set.Members {
		if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, member.Key); err != nil {
			return PackSet{}, err
		}
	}
	var current *string
	if err := tx.QueryRow(ctx, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&current); err != nil {
		return PackSet{}, err
	}
	if current == nil || before != *current {
		return PackSet{}, &OperationRejection{Reason: "stale_admission_state"}
	}
	err = insertPinTx(ctx, tx, set.ID, ownerKind, ownerID, operationID)
	return set, err
}

func insertPinTx(ctx context.Context, tx pgx.Tx, setID, ownerKind, ownerID string, operationID uuid.UUID) error {
	if ownerKind == "" || ownerID == "" || strings.ContainsRune(ownerKind, 0) || strings.ContainsRune(ownerID, 0) || operationID == uuid.Nil {
		return errors.New("reference pack: invalid pin owner")
	}
	_, err := tx.Exec(ctx, `INSERT INTO reference_pack_pins(owner_kind,owner_id,pack_set_id,operation_id) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, ownerKind, ownerID, setID, operationID)
	return err
}
