package recoverystate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

const (
	ContributionSchemaID = "cartulary.recovery_state_contribution.v1"
	CatalogSchemaID      = "cartulary.recovery_state_catalog.v1"
	PostgresUnitCodecID  = "cartulary.postgres_snapshot_unit.v1"

	AuthoredTableCount  = 114
	RequiredTableCount  = 84
	ContributionCount   = 30
	ObjectFamilyCount   = 6
	catalogFixturePath  = "contracts/recovery/fixtures/recovery-state-catalog.v1.json"
	catalogDigestPrefix = "CARTULARY-RECOVERY-STATE-CATALOG-V1\n"
)

var ErrInvalidCatalog = errors.New("recovery state catalog is invalid")

type StateClass string

const (
	StateAuthoritative StateClass = "authoritative"
	StateDerived       StateClass = "derived"
	StateTransient     StateClass = "transient"
	StateSecurity      StateClass = "security_ephemeral"
	StateRecovery      StateClass = "recovery_metadata"
	StateMigration     StateClass = "migration_metadata"
)

type BackupInclusion string

const (
	InclusionRequired         BackupInclusion = "authoritative_required"
	InclusionRebuildable      BackupInclusion = "excluded_rebuildable"
	InclusionTransient        BackupInclusion = "excluded_transient"
	InclusionSecurity         BackupInclusion = "excluded_security_state"
	InclusionRecoveryMetadata BackupInclusion = "excluded_recovery_metadata"
	InclusionSchemaMetadata   BackupInclusion = "excluded_schema_metadata"
)

type RestoreAction string

const (
	RestoreState    RestoreAction = "restore"
	RebuildState    RestoreAction = "rebuild"
	InvalidateState RestoreAction = "invalidate"
	IgnoreState     RestoreAction = "ignore"
)

type Table struct {
	TableName       string          `json:"table_name"`
	StateClass      StateClass      `json:"state_class"`
	BackupInclusion BackupInclusion `json:"backup_inclusion"`
	RestoreAction   RestoreAction   `json:"restore_action"`
	CodecID         *string         `json:"codec_id"`
	AlgorithmID     *string         `json:"algorithm_id"`
}

type ObjectFamily struct {
	ObjectFamilyID        string          `json:"object_family_id"`
	StateClass            StateClass      `json:"state_class"`
	BackupInclusion       BackupInclusion `json:"backup_inclusion"`
	RestoreAction         RestoreAction   `json:"restore_action"`
	InventoryAlgorithmID  *string         `json:"inventory_algorithm_id"`
	ValidationAlgorithmID *string         `json:"validation_algorithm_id"`
	RestoreAlgorithmID    *string         `json:"restore_algorithm_id"`
}

type Contribution struct {
	SchemaID       string         `json:"schema_id"`
	OwnerID        string         `json:"owner_id"`
	Tables         []Table        `json:"tables"`
	ObjectFamilies []ObjectFamily `json:"object_families"`
}

type ContributionDigest struct {
	OwnerID string `json:"owner_id"`
	SHA256  string `json:"sha256"`
}

type CatalogTable struct {
	OwnerID         string          `json:"owner_id"`
	TableName       string          `json:"table_name"`
	StateClass      StateClass      `json:"state_class"`
	BackupInclusion BackupInclusion `json:"backup_inclusion"`
	RestoreAction   RestoreAction   `json:"restore_action"`
	CodecID         *string         `json:"codec_id"`
	AlgorithmID     *string         `json:"algorithm_id"`
}

type CatalogObjectFamily struct {
	OwnerID               string          `json:"owner_id"`
	ObjectFamilyID        string          `json:"object_family_id"`
	StateClass            StateClass      `json:"state_class"`
	BackupInclusion       BackupInclusion `json:"backup_inclusion"`
	RestoreAction         RestoreAction   `json:"restore_action"`
	InventoryAlgorithmID  *string         `json:"inventory_algorithm_id"`
	ValidationAlgorithmID *string         `json:"validation_algorithm_id"`
	RestoreAlgorithmID    *string         `json:"restore_algorithm_id"`
}

type Document struct {
	SchemaID            string                `json:"schema_id"`
	CatalogDigestSHA256 string                `json:"catalog_digest_sha256"`
	ContributionDigests []ContributionDigest  `json:"contribution_digests"`
	Tables              []CatalogTable        `json:"tables"`
	ObjectFamilies      []CatalogObjectFamily `json:"object_families"`
}

type Catalog struct {
	document Document
	shape    FrozenShape
}

type FrozenShape struct {
	ContributionCount  int
	AuthoredTableCount int
	RequiredTableCount int
	ObjectFamilyCount  int
}

func AuthoritativeTables(names ...string) []Table {
	codec := PostgresUnitCodecID
	tables := make([]Table, 0, len(names))
	for _, name := range names {
		tables = append(tables, Table{
			TableName:       name,
			StateClass:      StateAuthoritative,
			BackupInclusion: InclusionRequired,
			RestoreAction:   RestoreState,
			CodecID:         &codec,
		})
	}
	return tables
}

func RebuildableTables(algorithmID string, names ...string) []Table {
	return algorithmTables(StateDerived, InclusionRebuildable, RebuildState, algorithmID, names...)
}

func SecurityStateTables(algorithmID string, names ...string) []Table {
	return algorithmTables(StateSecurity, InclusionSecurity, InvalidateState, algorithmID, names...)
}

// TransientStateTables describes restart-safe coordination state that is not
// restored. Restore invalidates it so owner reconciliation can reconstruct
// fresh coordination from authoritative state.
func TransientStateTables(algorithmID string, names ...string) []Table {
	return algorithmTables(StateTransient, InclusionTransient, InvalidateState, algorithmID, names...)
}

func RecoveryMetadataTables(names ...string) []Table {
	return ignoredTables(StateRecovery, InclusionRecoveryMetadata, names...)
}

func SchemaMetadataTables(names ...string) []Table {
	return ignoredTables(StateMigration, InclusionSchemaMetadata, names...)
}

func AuthoritativeObjectFamily(
	familyID string,
	inventoryAlgorithmID string,
	validationAlgorithmID string,
	restoreAlgorithmID string,
) ObjectFamily {
	return ObjectFamily{
		ObjectFamilyID:        familyID,
		StateClass:            StateAuthoritative,
		BackupInclusion:       InclusionRequired,
		RestoreAction:         RestoreState,
		InventoryAlgorithmID:  stringPointer(inventoryAlgorithmID),
		ValidationAlgorithmID: stringPointer(validationAlgorithmID),
		RestoreAlgorithmID:    stringPointer(restoreAlgorithmID),
	}
}

func NewContribution(ownerID string, tables []Table, objectFamilies ...ObjectFamily) Contribution {
	return Contribution{
		SchemaID:       ContributionSchemaID,
		OwnerID:        ownerID,
		Tables:         append([]Table(nil), tables...),
		ObjectFamilies: append([]ObjectFamily(nil), objectFamilies...),
	}
}

func Build(contributions ...Contribution) (*Catalog, error) {
	normalized := append([]Contribution(nil), contributions...)
	sort.Slice(normalized, func(left, right int) bool {
		return normalized[left].OwnerID < normalized[right].OwnerID
	})
	document := Document{
		SchemaID:            CatalogSchemaID,
		ContributionDigests: make([]ContributionDigest, 0, len(normalized)),
		Tables:              make([]CatalogTable, 0, AuthoredTableCount),
		ObjectFamilies:      make([]CatalogObjectFamily, 0, ObjectFamilyCount),
	}
	owners := make(map[string]struct{}, len(normalized))
	tableNames := make(map[string]string, AuthoredTableCount)
	objectFamilyIDs := make(map[string]string, ObjectFamilyCount)
	for index := range normalized {
		contribution := normalizeContribution(normalized[index])
		if err := validateContribution(contribution); err != nil {
			return nil, err
		}
		if _, duplicate := owners[contribution.OwnerID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate owner contribution %s", ErrInvalidCatalog, contribution.OwnerID)
		}
		owners[contribution.OwnerID] = struct{}{}
		digest, err := digestJSON(contribution)
		if err != nil {
			return nil, fmt.Errorf("%w: digest contribution %s: %v", ErrInvalidCatalog, contribution.OwnerID, err)
		}
		document.ContributionDigests = append(document.ContributionDigests, ContributionDigest{
			OwnerID: contribution.OwnerID,
			SHA256:  digest,
		})
		for _, table := range contribution.Tables {
			if previousOwner, duplicate := tableNames[table.TableName]; duplicate {
				return nil, fmt.Errorf(
					"%w: table %s is contributed by both %s and %s",
					ErrInvalidCatalog,
					table.TableName,
					previousOwner,
					contribution.OwnerID,
				)
			}
			tableNames[table.TableName] = contribution.OwnerID
			document.Tables = append(document.Tables, CatalogTable{
				OwnerID:         contribution.OwnerID,
				TableName:       table.TableName,
				StateClass:      table.StateClass,
				BackupInclusion: table.BackupInclusion,
				RestoreAction:   table.RestoreAction,
				CodecID:         cloneStringPointer(table.CodecID),
				AlgorithmID:     cloneStringPointer(table.AlgorithmID),
			})
		}
		for _, family := range contribution.ObjectFamilies {
			if previousOwner, duplicate := objectFamilyIDs[family.ObjectFamilyID]; duplicate {
				return nil, fmt.Errorf(
					"%w: object family %s is contributed by both %s and %s",
					ErrInvalidCatalog,
					family.ObjectFamilyID,
					previousOwner,
					contribution.OwnerID,
				)
			}
			objectFamilyIDs[family.ObjectFamilyID] = contribution.OwnerID
			document.ObjectFamilies = append(document.ObjectFamilies, CatalogObjectFamily{
				OwnerID:               contribution.OwnerID,
				ObjectFamilyID:        family.ObjectFamilyID,
				StateClass:            family.StateClass,
				BackupInclusion:       family.BackupInclusion,
				RestoreAction:         family.RestoreAction,
				InventoryAlgorithmID:  cloneStringPointer(family.InventoryAlgorithmID),
				ValidationAlgorithmID: cloneStringPointer(family.ValidationAlgorithmID),
				RestoreAlgorithmID:    cloneStringPointer(family.RestoreAlgorithmID),
			})
		}
	}
	sort.Slice(document.Tables, func(left, right int) bool {
		return document.Tables[left].TableName < document.Tables[right].TableName
	})
	sort.Slice(document.ObjectFamilies, func(left, right int) bool {
		return document.ObjectFamilies[left].ObjectFamilyID < document.ObjectFamilies[right].ObjectFamilyID
	})
	shape := FrozenShape{
		ContributionCount: ContributionCount, AuthoredTableCount: AuthoredTableCount,
		RequiredTableCount: RequiredTableCount, ObjectFamilyCount: ObjectFamilyCount,
	}
	if err := validateCurrentDocument(document, shape); err != nil {
		return nil, err
	}
	digest, err := calculateDocumentDigest(document)
	if err != nil {
		return nil, err
	}
	document.CatalogDigestSHA256 = digest
	return &Catalog{document: cloneDocument(document), shape: shape}, nil
}

func NewFrozenCatalogJSON(body []byte, shape FrozenShape) (*Catalog, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("%w: decode frozen catalog: %v", ErrInvalidCatalog, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("%w: frozen catalog has trailing JSON", ErrInvalidCatalog)
	}
	catalog := &Catalog{document: cloneDocument(document), shape: shape}
	if err := catalog.ValidateFrozen(); err != nil {
		return nil, err
	}
	return catalog, nil
}

func (catalog *Catalog) ValidateFrozen() error {
	if catalog == nil {
		return fmt.Errorf("%w: frozen catalog is unavailable", ErrInvalidCatalog)
	}
	if err := validateFrozenDocument(catalog.document, catalog.shape); err != nil {
		return err
	}
	digest, err := calculateDocumentDigest(catalog.document)
	if err != nil {
		return err
	}
	if digest != catalog.document.CatalogDigestSHA256 {
		return fmt.Errorf("%w: frozen catalog digest mismatch", ErrInvalidCatalog)
	}
	return nil
}

func (catalog *Catalog) Document() Document {
	if catalog == nil {
		return Document{}
	}
	return cloneDocument(catalog.document)
}

func (catalog *Catalog) DigestSHA256() string {
	if catalog == nil {
		return ""
	}
	return catalog.document.CatalogDigestSHA256
}

func (catalog *Catalog) RequiredTableNames() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, 0, catalog.shape.RequiredTableCount)
	for _, table := range catalog.document.Tables {
		if table.BackupInclusion == InclusionRequired {
			result = append(result, table.TableName)
		}
	}
	return result
}

func (catalog *Catalog) AuthoredTableNames() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, 0, len(catalog.document.Tables))
	for _, table := range catalog.document.Tables {
		result = append(result, table.TableName)
	}
	return result
}

func (catalog *Catalog) ValidateDatabaseTableNames(actual []string) error {
	if err := catalog.ValidateFrozen(); err != nil {
		return err
	}
	expected := catalog.AuthoredTableNames()
	expected = append(expected, "goose_db_version")
	sort.Strings(expected)
	actual = append([]string(nil), actual...)
	sort.Strings(actual)
	return compareStrings(expected, actual, "database tables and frozen catalog")
}

func (catalog *Catalog) ValidateLegacyShadowTables(actual []string) error {
	if catalog == nil {
		return fmt.Errorf("%w: frozen catalog is unavailable", ErrInvalidCatalog)
	}
	expected := catalog.RequiredTableNames()
	expected = append(expected,
		"collaboration_event_intents",
		"collaboration_incident_stream_cursors",
		"collaboration_replay_events",
		"collaboration_resume_tokens",
		"enterprise_auth_transactions",
	)
	sort.Strings(expected)
	actual = append([]string(nil), actual...)
	sort.Strings(actual)
	return compareStrings(expected, actual, "legacy snapshot and catalog shadow")
}

func algorithmTables(
	stateClass StateClass,
	inclusion BackupInclusion,
	action RestoreAction,
	algorithmID string,
	names ...string,
) []Table {
	tables := make([]Table, 0, len(names))
	for _, name := range names {
		tables = append(tables, Table{
			TableName:       name,
			StateClass:      stateClass,
			BackupInclusion: inclusion,
			RestoreAction:   action,
			AlgorithmID:     stringPointer(algorithmID),
		})
	}
	return tables
}

func ignoredTables(stateClass StateClass, inclusion BackupInclusion, names ...string) []Table {
	tables := make([]Table, 0, len(names))
	for _, name := range names {
		tables = append(tables, Table{
			TableName:       name,
			StateClass:      stateClass,
			BackupInclusion: inclusion,
			RestoreAction:   IgnoreState,
		})
	}
	return tables
}

func normalizeContribution(contribution Contribution) Contribution {
	normalized := Contribution{
		SchemaID:       contribution.SchemaID,
		OwnerID:        strings.TrimSpace(contribution.OwnerID),
		Tables:         append([]Table(nil), contribution.Tables...),
		ObjectFamilies: append([]ObjectFamily(nil), contribution.ObjectFamilies...),
	}
	sort.Slice(normalized.Tables, func(left, right int) bool {
		return normalized.Tables[left].TableName < normalized.Tables[right].TableName
	})
	sort.Slice(normalized.ObjectFamilies, func(left, right int) bool {
		return normalized.ObjectFamilies[left].ObjectFamilyID < normalized.ObjectFamilies[right].ObjectFamilyID
	})
	return normalized
}

func validateContribution(contribution Contribution) error {
	if contribution.SchemaID != ContributionSchemaID {
		return fmt.Errorf("%w: owner %s has wrong contribution schema", ErrInvalidCatalog, contribution.OwnerID)
	}
	if !validIdentifier(contribution.OwnerID) {
		return fmt.Errorf("%w: invalid owner_id %q", ErrInvalidCatalog, contribution.OwnerID)
	}
	seenTables := make(map[string]struct{}, len(contribution.Tables))
	for _, table := range contribution.Tables {
		if _, duplicate := seenTables[table.TableName]; duplicate {
			return fmt.Errorf("%w: owner %s duplicates table %s", ErrInvalidCatalog, contribution.OwnerID, table.TableName)
		}
		seenTables[table.TableName] = struct{}{}
		if err := validateTable(table); err != nil {
			return fmt.Errorf("%w: owner %s: %v", ErrInvalidCatalog, contribution.OwnerID, err)
		}
	}
	seenFamilies := make(map[string]struct{}, len(contribution.ObjectFamilies))
	for _, family := range contribution.ObjectFamilies {
		if _, duplicate := seenFamilies[family.ObjectFamilyID]; duplicate {
			return fmt.Errorf("%w: owner %s duplicates object family %s", ErrInvalidCatalog, contribution.OwnerID, family.ObjectFamilyID)
		}
		seenFamilies[family.ObjectFamilyID] = struct{}{}
		if err := validateObjectFamily(family); err != nil {
			return fmt.Errorf("%w: owner %s: %v", ErrInvalidCatalog, contribution.OwnerID, err)
		}
	}
	return nil
}

func validateTable(table Table) error {
	if !validTableName(table.TableName) {
		return fmt.Errorf("invalid table_name %q", table.TableName)
	}
	switch table.BackupInclusion {
	case InclusionRequired:
		if table.StateClass != StateAuthoritative || table.RestoreAction != RestoreState ||
			stringValue(table.CodecID) != PostgresUnitCodecID || table.AlgorithmID != nil {
			return fmt.Errorf("authoritative table %s has inconsistent restore facts", table.TableName)
		}
	case InclusionRebuildable:
		if table.StateClass != StateDerived || table.RestoreAction != RebuildState ||
			table.CodecID != nil || !validIdentifier(stringValue(table.AlgorithmID)) {
			return fmt.Errorf("rebuildable table %s has inconsistent restore facts", table.TableName)
		}
	case InclusionSecurity:
		if table.StateClass != StateSecurity || table.RestoreAction != InvalidateState ||
			table.CodecID != nil || !validIdentifier(stringValue(table.AlgorithmID)) {
			return fmt.Errorf("security table %s has inconsistent invalidation facts", table.TableName)
		}
	case InclusionRecoveryMetadata:
		if table.StateClass != StateRecovery || table.RestoreAction != IgnoreState ||
			table.CodecID != nil || table.AlgorithmID != nil {
			return fmt.Errorf("recovery metadata table %s has inconsistent exclusion facts", table.TableName)
		}
	case InclusionSchemaMetadata:
		if table.StateClass != StateMigration || table.RestoreAction != IgnoreState ||
			table.CodecID != nil || table.AlgorithmID != nil {
			return fmt.Errorf("schema metadata table %s has inconsistent exclusion facts", table.TableName)
		}
	case InclusionTransient:
		if table.StateClass != StateTransient || table.RestoreAction != InvalidateState ||
			table.CodecID != nil || !validIdentifier(stringValue(table.AlgorithmID)) {
			return fmt.Errorf("transient table %s has inconsistent exclusion facts", table.TableName)
		}
	default:
		return fmt.Errorf("table %s has unsupported backup inclusion %q", table.TableName, table.BackupInclusion)
	}
	return nil
}

func validateObjectFamily(family ObjectFamily) error {
	if !validIdentifier(family.ObjectFamilyID) {
		return fmt.Errorf("invalid object_family_id %q", family.ObjectFamilyID)
	}
	if family.StateClass != StateAuthoritative ||
		family.BackupInclusion != InclusionRequired ||
		family.RestoreAction != RestoreState ||
		!validIdentifier(stringValue(family.InventoryAlgorithmID)) ||
		!validIdentifier(stringValue(family.ValidationAlgorithmID)) ||
		!validIdentifier(stringValue(family.RestoreAlgorithmID)) {
		return fmt.Errorf("object family %s has incomplete authoritative algorithms", family.ObjectFamilyID)
	}
	return nil
}

func validateCurrentDocument(document Document, shape FrozenShape) error {
	if err := validateFrozenDocument(document, shape); err != nil {
		return err
	}
	expectedArtifact, ok := contractrecovery.Index[catalogFixturePath]
	if !ok {
		return fmt.Errorf("%w: generated catalog projection is unavailable", ErrInvalidCatalog)
	}
	var expected Document
	if err := json.Unmarshal([]byte(expectedArtifact.JSON), &expected); err != nil {
		return fmt.Errorf("%w: decode generated catalog projection: %v", ErrInvalidCatalog, err)
	}
	if err := compareContributionDigests(expected.ContributionDigests, document.ContributionDigests); err != nil {
		return err
	}
	if err := compareCatalogTables(expected.Tables, document.Tables); err != nil {
		return err
	}
	return compareCatalogObjectFamilies(expected.ObjectFamilies, document.ObjectFamilies)
}

func validateFrozenDocument(document Document, shape FrozenShape) error {
	if document.SchemaID != CatalogSchemaID {
		return fmt.Errorf("%w: wrong catalog schema", ErrInvalidCatalog)
	}
	if shape.ContributionCount <= 0 || shape.AuthoredTableCount <= 0 ||
		shape.RequiredTableCount <= 0 || shape.ObjectFamilyCount <= 0 {
		return fmt.Errorf("%w: frozen catalog shape is invalid", ErrInvalidCatalog)
	}
	if len(document.ContributionDigests) != shape.ContributionCount {
		return fmt.Errorf(
			"%w: expected %d owner contributions, got %d",
			ErrInvalidCatalog,
			shape.ContributionCount,
			len(document.ContributionDigests),
		)
	}
	previousOwner := ""
	for _, contribution := range document.ContributionDigests {
		if !validIdentifier(contribution.OwnerID) || !validSHA256Hex(contribution.SHA256) ||
			(previousOwner != "" && contribution.OwnerID <= previousOwner) {
			return fmt.Errorf("%w: contribution digests are malformed, duplicated, or unsorted", ErrInvalidCatalog)
		}
		previousOwner = contribution.OwnerID
	}
	if len(document.Tables) != shape.AuthoredTableCount {
		return fmt.Errorf(
			"%w: expected %d authored tables, got %d",
			ErrInvalidCatalog,
			shape.AuthoredTableCount,
			len(document.Tables),
		)
	}
	required := 0
	previousTable := ""
	for _, table := range document.Tables {
		if !validIdentifier(table.OwnerID) || (previousTable != "" && table.TableName <= previousTable) {
			return fmt.Errorf("%w: catalog tables are malformed, duplicated, or unsorted", ErrInvalidCatalog)
		}
		if err := validateTable(Table{
			TableName: table.TableName, StateClass: table.StateClass,
			BackupInclusion: table.BackupInclusion, RestoreAction: table.RestoreAction,
			CodecID: table.CodecID, AlgorithmID: table.AlgorithmID,
		}); err != nil {
			return fmt.Errorf("%w: owner %s: %v", ErrInvalidCatalog, table.OwnerID, err)
		}
		previousTable = table.TableName
		if table.BackupInclusion == InclusionRequired {
			required++
		}
	}
	if required != shape.RequiredTableCount {
		return fmt.Errorf(
			"%w: expected %d required tables, got %d",
			ErrInvalidCatalog,
			shape.RequiredTableCount,
			required,
		)
	}
	if len(document.ObjectFamilies) != shape.ObjectFamilyCount {
		return fmt.Errorf(
			"%w: expected %d object families, got %d",
			ErrInvalidCatalog,
			shape.ObjectFamilyCount,
			len(document.ObjectFamilies),
		)
	}
	previousFamily := ""
	for _, family := range document.ObjectFamilies {
		if !validIdentifier(family.OwnerID) ||
			(previousFamily != "" && family.ObjectFamilyID <= previousFamily) {
			return fmt.Errorf("%w: catalog object families are malformed, duplicated, or unsorted", ErrInvalidCatalog)
		}
		if err := validateObjectFamily(ObjectFamily{
			ObjectFamilyID: family.ObjectFamilyID, StateClass: family.StateClass,
			BackupInclusion: family.BackupInclusion, RestoreAction: family.RestoreAction,
			InventoryAlgorithmID:  family.InventoryAlgorithmID,
			ValidationAlgorithmID: family.ValidationAlgorithmID,
			RestoreAlgorithmID:    family.RestoreAlgorithmID,
		}); err != nil {
			return fmt.Errorf("%w: owner %s: %v", ErrInvalidCatalog, family.OwnerID, err)
		}
		previousFamily = family.ObjectFamilyID
	}
	return nil
}

func compareContributionDigests(expected []ContributionDigest, actual []ContributionDigest) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%w: generated and assembled contribution counts differ", ErrInvalidCatalog)
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return fmt.Errorf("%w: generated and assembled contributions differ at owner %s", ErrInvalidCatalog, expected[index].OwnerID)
		}
	}
	return nil
}

func compareCatalogTables(expected []CatalogTable, actual []CatalogTable) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("%w: generated and assembled catalog table counts differ", ErrInvalidCatalog)
	}
	for index := range expected {
		if !catalogTablesEqual(expected[index], actual[index]) {
			return fmt.Errorf(
				"%w: generated and assembled catalog differ at table %s",
				ErrInvalidCatalog,
				expected[index].TableName,
			)
		}
	}
	return nil
}

func compareCatalogObjectFamilies(expected []CatalogObjectFamily, actual []CatalogObjectFamily) error {
	sort.Slice(expected, func(left, right int) bool {
		return expected[left].ObjectFamilyID < expected[right].ObjectFamilyID
	})
	if len(expected) != len(actual) {
		return fmt.Errorf("%w: generated and assembled object family counts differ", ErrInvalidCatalog)
	}
	for index := range expected {
		if !catalogObjectFamiliesEqual(expected[index], actual[index]) {
			return fmt.Errorf(
				"%w: generated and assembled catalog differ at object family %s",
				ErrInvalidCatalog,
				expected[index].ObjectFamilyID,
			)
		}
	}
	return nil
}

func catalogTablesEqual(left CatalogTable, right CatalogTable) bool {
	return left.OwnerID == right.OwnerID &&
		left.TableName == right.TableName &&
		left.StateClass == right.StateClass &&
		left.BackupInclusion == right.BackupInclusion &&
		left.RestoreAction == right.RestoreAction &&
		stringValue(left.CodecID) == stringValue(right.CodecID) &&
		stringValue(left.AlgorithmID) == stringValue(right.AlgorithmID)
}

func catalogObjectFamiliesEqual(left CatalogObjectFamily, right CatalogObjectFamily) bool {
	return left.OwnerID == right.OwnerID &&
		left.ObjectFamilyID == right.ObjectFamilyID &&
		left.StateClass == right.StateClass &&
		left.BackupInclusion == right.BackupInclusion &&
		left.RestoreAction == right.RestoreAction &&
		stringValue(left.InventoryAlgorithmID) == stringValue(right.InventoryAlgorithmID) &&
		stringValue(left.ValidationAlgorithmID) == stringValue(right.ValidationAlgorithmID) &&
		stringValue(left.RestoreAlgorithmID) == stringValue(right.RestoreAlgorithmID)
}

func compareStrings(expected []string, actual []string, label string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf(
			"%w: %s counts differ: expected %d, got %d",
			ErrInvalidCatalog,
			label,
			len(expected),
			len(actual),
		)
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return fmt.Errorf(
				"%w: %s differ at index %d: expected %s, got %s",
				ErrInvalidCatalog,
				label,
				index,
				expected[index],
				actual[index],
			)
		}
	}
	return nil
}

func digestJSON(value any) (string, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func calculateDocumentDigest(document Document) (string, error) {
	digestPreimage := struct {
		SchemaID            string                `json:"schema_id"`
		ContributionDigests []ContributionDigest  `json:"contribution_digests"`
		Tables              []CatalogTable        `json:"tables"`
		ObjectFamilies      []CatalogObjectFamily `json:"object_families"`
	}{
		SchemaID:            document.SchemaID,
		ContributionDigests: document.ContributionDigests,
		Tables:              document.Tables,
		ObjectFamilies:      document.ObjectFamilies,
	}
	canonical, err := json.Marshal(digestPreimage)
	if err != nil {
		return "", fmt.Errorf("%w: encode catalog digest preimage: %v", ErrInvalidCatalog, err)
	}
	sum := sha256.Sum256(append([]byte(catalogDigestPrefix), canonical...))
	return hex.EncodeToString(sum[:]), nil
}

func cloneDocument(document Document) Document {
	cloned := document
	cloned.ContributionDigests = append([]ContributionDigest(nil), document.ContributionDigests...)
	cloned.Tables = append([]CatalogTable(nil), document.Tables...)
	for index := range cloned.Tables {
		cloned.Tables[index].CodecID = cloneStringPointer(cloned.Tables[index].CodecID)
		cloned.Tables[index].AlgorithmID = cloneStringPointer(cloned.Tables[index].AlgorithmID)
	}
	cloned.ObjectFamilies = append([]CatalogObjectFamily(nil), document.ObjectFamilies...)
	for index := range cloned.ObjectFamilies {
		cloned.ObjectFamilies[index].InventoryAlgorithmID = cloneStringPointer(cloned.ObjectFamilies[index].InventoryAlgorithmID)
		cloned.ObjectFamilies[index].ValidationAlgorithmID = cloneStringPointer(cloned.ObjectFamilies[index].ValidationAlgorithmID)
		cloned.ObjectFamilies[index].RestoreAlgorithmID = cloneStringPointer(cloned.ObjectFamilies[index].RestoreAlgorithmID)
	}
	return cloned
}

func stringPointer(value string) *string {
	return &value
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validTableName(value string) bool {
	if value == "" || len(value) > 63 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			character != '_' {
			return false
		}
	}
	return true
}

func validIdentifier(value string) bool {
	if value == "" || len(value) > 160 {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') &&
			character != '_' &&
			character != '.' &&
			character != ':' &&
			character != '-' {
			return false
		}
	}
	first := value[0]
	return (first >= 'a' && first <= 'z') ||
		(first >= 'A' && first <= 'Z') ||
		(first >= '0' && first <= '9')
}

func validSHA256Hex(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}
