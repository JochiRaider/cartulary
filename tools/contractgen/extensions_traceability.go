package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const extensionDocumentAnchorID = "cartulary.extensions_subsystem_nlspec.v1"

var (
	extensionRequirementMarker = regexp.MustCompile(`^\*\*(EXT-REQ-[0-9]{3})\*\*$`)
	extensionAcceptanceID      = regexp.MustCompile(`EXT-AC-[0-9]{3}`)
	extensionH1                = regexp.MustCompile(`^# ([0-9]+)(?:\.|\s)`)
	extensionListItem          = regexp.MustCompile(`^(?:- |[0-9]+\. )`)
	extensionTableCaption      = regexp.MustCompile(`^\*\*Table [^*]+\*\*$`)
	extensionATXHeading        = regexp.MustCompile(`^(#{1,6}) `)
)

type extensionSourceLine struct {
	start int
	end   int
	text  string
}

type extensionExtractedClause struct {
	start            int
	end              int
	parentAnchorKind string
	parentAnchorID   string
	clauseKind       string
	requirementIDs   []string
	acceptanceIDs    []string
	verificationIDs  []string
}

type extensionRequirementTrace struct {
	sectionID     string
	acceptanceIDs []string
}

type extensionSourceRange struct {
	start int
	end   int
}

func extensionDocumentLines(document []byte) []extensionSourceLine {
	lines := make([]extensionSourceLine, 0, bytes.Count(document, []byte{'\n'}))
	start := 0
	for start < len(document) {
		relativeEnd := bytes.IndexByte(document[start:], '\n')
		end := len(document)
		if relativeEnd >= 0 {
			end = start + relativeEnd + 1
		}
		text := strings.TrimSuffix(string(document[start:end]), "\n")
		lines = append(lines, extensionSourceLine{start: start, end: end, text: text})
		start = end
	}
	return lines
}

func lintExtensionNormativeSource(document []byte) []string {
	findings := []string{}
	add := func(format string, args ...any) { findings = append(findings, fmt.Sprintf(format, args...)) }
	if !utf8.Valid(document) {
		add("source is not valid UTF-8")
	}
	if bytes.HasPrefix(document, []byte{0xef, 0xbb, 0xbf}) {
		add("source has a UTF-8 BOM")
	}
	for _, invalid := range []struct {
		value byte
		name  string
	}{{'\r', "CR"}, {0, "NUL"}, {'\t', "tab"}} {
		if bytes.IndexByte(document, invalid.value) >= 0 {
			add("source contains %s", invalid.name)
		}
	}
	if len(document) == 0 || document[len(document)-1] != '\n' || (len(document) > 1 && document[len(document)-2] == '\n') {
		add("source must end with exactly one LF")
	}

	lines := extensionDocumentLines(document)
	headingDepth := 0
	inFence := false
	tableColumns := 0
	seenRequirements := map[string]struct{}{}
	seenAcceptance := map[string]struct{}{}
	frontmatter := false
	for index, line := range lines {
		text := line.text
		if index == 0 && text == "---" {
			frontmatter = true
			continue
		}
		if frontmatter && text == "---" {
			frontmatter = false
			continue
		}
		if strings.HasPrefix(text, "````") {
			add("line %d uses a non-three-backtick fence", index+1)
			continue
		}
		if strings.HasPrefix(text, "```") {
			if inFence {
				if text != "```" {
					add("line %d has an invalid closing fence", index+1)
				}
				inFence = false
			} else {
				inFence = true
			}
			continue
		}
		if inFence {
			continue
		}
		if strings.HasPrefix(text, "    ") && strings.TrimSpace(text) != "" {
			add("line %d is indented code or nested content", index+1)
		}
		if strings.HasPrefix(text, "<") && strings.HasSuffix(text, ">") {
			add("line %d is a raw HTML block", index+1)
		}
		if match := extensionATXHeading.FindStringSubmatch(text); match != nil {
			depth := len(match[1])
			if headingDepth != 0 && depth > headingDepth+1 {
				add("line %d skips heading depth", index+1)
			}
			headingDepth = depth
		} else if strings.HasPrefix(text, "#") {
			add("line %d has invalid ATX heading syntax", index+1)
		}
		if !frontmatter && index+1 < len(lines) && text != "" && lines[index+1].text != "---" && regexp.MustCompile(`^(?:=+|-+)$`).MatchString(lines[index+1].text) {
			add("line %d uses a Setext heading", index+1)
		}
		if match := extensionRequirementMarker.FindStringSubmatch(text); match != nil {
			if _, duplicate := seenRequirements[match[1]]; duplicate {
				add("duplicate requirement %s", match[1])
			}
			seenRequirements[match[1]] = struct{}{}
		}
		if strings.HasPrefix(text, "|") {
			if !strings.HasSuffix(text, "|") {
				add("line %d has an open pipe table row", index+1)
			}
			columns := extensionPipeColumnCount(text)
			if tableColumns == 0 {
				tableColumns = columns
				if index+1 >= len(lines) || !extensionTableDelimiter(lines[index+1].text) {
					add("line %d table header has no exact delimiter row", index+1)
				}
			} else if columns != tableColumns {
				add("line %d has %d table columns; want %d", index+1, columns, tableColumns)
			}
		} else if text == "" {
			tableColumns = 0
		}
		if strings.HasPrefix(text, "| `EXT-AC-") {
			match := extensionAcceptanceID.FindString(text)
			if match != "" {
				if _, duplicate := seenAcceptance[match]; duplicate {
					add("duplicate acceptance criterion %s", match)
				}
				seenAcceptance[match] = struct{}{}
			}
		}
	}
	if inFence {
		add("source has an unterminated fence")
	}
	for value := 1; value <= 158; value++ {
		id := fmt.Sprintf("EXT-AC-%03d", value)
		if _, ok := seenAcceptance[id]; !ok {
			add("acceptance criteria omit %s", id)
		}
	}
	if len(seenAcceptance) != 158 {
		add("acceptance criteria contain %d unique rows; want 158", len(seenAcceptance))
	}
	if len(seenRequirements) != 236 {
		add("requirements contain %d unique markers; want 236", len(seenRequirements))
	}
	sort.Strings(findings)
	return findings
}

func extensionPipeColumnCount(text string) int {
	count := 0
	escaped := false
	for _, current := range text {
		if current == '\\' && !escaped {
			escaped = true
			continue
		}
		if current == '|' && !escaped {
			count++
		}
		escaped = false
	}
	if strings.HasPrefix(text, "|") && strings.HasSuffix(text, "|") {
		return count - 1
	}
	return count
}

func extensionTableDelimiter(text string) bool {
	if !strings.HasPrefix(text, "|") || !strings.HasSuffix(text, "|") {
		return false
	}
	for _, cell := range strings.Split(strings.Trim(text, "|"), "|") {
		trimmed := strings.TrimSpace(cell)
		trimmed = strings.TrimPrefix(trimmed, ":")
		trimmed = strings.TrimSuffix(trimmed, ":")
		if len(trimmed) < 3 || strings.Trim(trimmed, "-") != "" {
			return false
		}
	}
	return true
}

func extensionRequirementTraces(lines []extensionSourceLine) (map[string]extensionRequirementTrace, map[string][]string, error) {
	requirements := map[string]extensionRequirementTrace{}
	sectionRequirements := map[string][]string{}
	sectionID := ""
	currentRequirement := ""
	for _, line := range lines {
		if match := extensionH1.FindStringSubmatch(line.text); match != nil {
			sectionID = match[1]
			currentRequirement = ""
			continue
		}
		if match := extensionRequirementMarker.FindStringSubmatch(line.text); match != nil {
			currentRequirement = match[1]
			requirements[currentRequirement] = extensionRequirementTrace{sectionID: sectionID}
			sectionRequirements[sectionID] = append(sectionRequirements[sectionID], currentRequirement)
			continue
		}
		if currentRequirement != "" && strings.HasPrefix(line.text, "Verified by: ") {
			matches := extensionAcceptanceID.FindAllString(line.text, -1)
			if len(matches) == 0 {
				return nil, nil, fmt.Errorf("%s has an empty Verified by line", currentRequirement)
			}
			trace := requirements[currentRequirement]
			trace.acceptanceIDs = sortedExtensionIDs(matches)
			requirements[currentRequirement] = trace
		}
	}
	if len(requirements) != 236 {
		return nil, nil, fmt.Errorf("found %d extension requirements; want 236", len(requirements))
	}
	for requirementID, trace := range requirements {
		if len(trace.acceptanceIDs) == 0 {
			return nil, nil, fmt.Errorf("%s has no acceptance mapping", requirementID)
		}
	}
	return requirements, sectionRequirements, nil
}

func extractExtensionClauses(document []byte) ([]extensionExtractedClause, error) {
	if findings := lintExtensionNormativeSource(document); len(findings) != 0 {
		return nil, fmt.Errorf("extension normative source is invalid: %s", strings.Join(findings, "; "))
	}
	appendixOffset := bytes.Index(document, []byte("# Appendix A."))
	if appendixOffset < 0 {
		return nil, fmt.Errorf("extension source omits Appendix A boundary")
	}
	lines := extensionDocumentLines(document[:appendixOffset])
	requirements, sectionRequirements, err := extensionRequirementTraces(lines)
	if err != nil {
		return nil, err
	}
	reverseRequirements := map[string][]string{}
	for requirementID, trace := range requirements {
		for _, acceptanceID := range trace.acceptanceIDs {
			reverseRequirements[acceptanceID] = append(reverseRequirements[acceptanceID], requirementID)
		}
	}
	for value := 1; value <= 158; value++ {
		acceptanceID := fmt.Sprintf("EXT-AC-%03d", value)
		if len(reverseRequirements[acceptanceID]) == 0 {
			return nil, fmt.Errorf("%s has no requirement mapping", acceptanceID)
		}
		reverseRequirements[acceptanceID] = sortedExtensionIDs(reverseRequirements[acceptanceID])
	}

	clauses := []extensionExtractedClause{}
	frontmatter := false
	frontmatterClosed := false
	sectionID := ""
	parentKind := "document"
	parentID := extensionDocumentAnchorID
	inFence := false
	for index := 0; index < len(lines); {
		line := lines[index]
		text := line.text
		if index == 0 && text == "---" {
			frontmatter = true
			index++
			continue
		}
		if frontmatter {
			if text == "---" {
				frontmatter = false
				frontmatterClosed = true
				index++
				continue
			}
			clauses = append(clauses, extensionExtractedClause{
				start: line.start, end: line.end, parentAnchorKind: "document", parentAnchorID: extensionDocumentAnchorID,
				clauseKind: "frontmatter_member", requirementIDs: []string{"EXT-REQ-168"}, acceptanceIDs: []string{"EXT-AC-001", "EXT-AC-072", "EXT-AC-075", "EXT-AC-128"},
				verificationIDs: []string{"module.extensions.verification.contract_accounting"},
			})
			index++
			continue
		}
		if !frontmatterClosed {
			return nil, fmt.Errorf("extension source has no closed front matter")
		}
		if match := extensionH1.FindStringSubmatch(text); match != nil {
			sectionID = match[1]
			parentKind, parentID = "h1", sectionID
			index++
			continue
		}
		if extensionATXHeading.MatchString(text) || text == "" || strings.HasPrefix(text, "Profiles:") || strings.HasPrefix(text, "Verified by:") {
			index++
			continue
		}
		if match := extensionRequirementMarker.FindStringSubmatch(text); match != nil {
			parentKind, parentID = "requirement", match[1]
			index++
			continue
		}
		if strings.HasPrefix(text, "```") {
			start := line.start
			index++
			for index < len(lines) && lines[index].text != "```" {
				index++
			}
			if index >= len(lines) {
				return nil, fmt.Errorf("unterminated fenced literal at byte %d", start)
			}
			end := lines[index].end
			index++
			clause := extensionExtractedClause{start: start, end: end, parentAnchorKind: parentKind, parentAnchorID: parentID, clauseKind: "fenced_literal"}
			assignExtensionClauseTrace(&clause, sectionID, requirements, sectionRequirements, reverseRequirements, document[start:end])
			clauses = append(clauses, clause)
			inFence = false
			continue
		}
		if inFence {
			return nil, fmt.Errorf("unexpected fence state")
		}
		kind := ""
		if extensionTableCaption.MatchString(text) {
			kind = "normative_table_caption"
		} else if strings.HasPrefix(text, "|") {
			if extensionTableDelimiter(text) || (index+1 < len(lines) && extensionTableDelimiter(lines[index+1].text)) {
				index++
				continue
			}
			kind = "normative_table_row"
			if sectionID == "28" && strings.HasPrefix(text, "| `EXT-AC-") {
				kind = "acceptance_row"
			}
		} else if extensionListItem.MatchString(text) {
			kind = "list_item"
		}
		if kind != "" {
			clause := extensionExtractedClause{start: line.start, end: line.end, parentAnchorKind: parentKind, parentAnchorID: parentID, clauseKind: kind}
			assignExtensionClauseTrace(&clause, sectionID, requirements, sectionRequirements, reverseRequirements, document[line.start:line.end])
			clauses = append(clauses, clause)
			index++
			continue
		}

		start := line.start
		end := line.end
		index++
		for index < len(lines) && !extensionStructuralLine(lines, index) {
			end = lines[index].end
			index++
		}
		clause := extensionExtractedClause{start: start, end: end, parentAnchorKind: parentKind, parentAnchorID: parentID, clauseKind: "prose_block"}
		assignExtensionClauseTrace(&clause, sectionID, requirements, sectionRequirements, reverseRequirements, document[start:end])
		clauses = append(clauses, clause)
	}
	return clauses, nil
}

func extensionStructuralLine(lines []extensionSourceLine, index int) bool {
	text := lines[index].text
	return text == "" || strings.HasPrefix(text, "```") || extensionATXHeading.MatchString(text) ||
		extensionRequirementMarker.MatchString(text) || strings.HasPrefix(text, "Profiles:") ||
		strings.HasPrefix(text, "Verified by:") || extensionTableCaption.MatchString(text) ||
		strings.HasPrefix(text, "|") || extensionListItem.MatchString(text)
}

func assignExtensionClauseTrace(clause *extensionExtractedClause, sectionID string, requirements map[string]extensionRequirementTrace, sectionRequirements map[string][]string, reverseRequirements map[string][]string, clauseBytes []byte) {
	if clause.clauseKind == "acceptance_row" {
		acceptanceID := extensionAcceptanceID.FindString(string(clauseBytes))
		clause.acceptanceIDs = []string{acceptanceID}
		clause.requirementIDs = append([]string(nil), reverseRequirements[acceptanceID]...)
	} else if clause.parentAnchorKind == "requirement" {
		clause.requirementIDs = []string{clause.parentAnchorID}
		clause.acceptanceIDs = append([]string(nil), requirements[clause.parentAnchorID].acceptanceIDs...)
	} else {
		clause.requirementIDs = append([]string(nil), sectionRequirements[sectionID]...)
		if len(clause.requirementIDs) == 0 && sectionID == "28" {
			clause.requirementIDs = []string{"EXT-REQ-166", "EXT-REQ-202", "EXT-REQ-236"}
		}
		for _, requirementID := range clause.requirementIDs {
			clause.acceptanceIDs = append(clause.acceptanceIDs, requirements[requirementID].acceptanceIDs...)
		}
		clause.requirementIDs = sortedExtensionIDs(clause.requirementIDs)
		clause.acceptanceIDs = sortedExtensionIDs(clause.acceptanceIDs)
	}
	clause.verificationIDs = extensionVerificationIDs(clause.requirementIDs)
}

func extensionVerificationIDs(requirementIDs []string) []string {
	accounting := map[string]struct{}{
		"EXT-REQ-165": {}, "EXT-REQ-166": {}, "EXT-REQ-167": {}, "EXT-REQ-168": {},
		"EXT-REQ-174": {}, "EXT-REQ-202": {}, "EXT-REQ-224": {}, "EXT-REQ-226": {},
		"EXT-REQ-228": {}, "EXT-REQ-229": {}, "EXT-REQ-230": {}, "EXT-REQ-236": {},
	}
	includeBehavior := false
	includeAccounting := false
	for _, requirementID := range requirementIDs {
		if _, ok := accounting[requirementID]; ok {
			includeAccounting = true
		} else {
			includeBehavior = true
		}
	}
	verificationIDs := []string{}
	if includeBehavior {
		verificationIDs = append(verificationIDs, "module.extensions.verification.behavior_contract")
	}
	if includeAccounting {
		verificationIDs = append(verificationIDs, "module.extensions.verification.contract_accounting")
	}
	return verificationIDs
}

func sortedExtensionIDs(values []string) []string {
	set := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		leftPrefix, leftNumber := extensionIDParts(result[i])
		rightPrefix, rightNumber := extensionIDParts(result[j])
		if leftPrefix != rightPrefix {
			return leftPrefix < rightPrefix
		}
		return leftNumber < rightNumber
	})
	return result
}

func extensionIDParts(value string) (string, int) {
	lastDash := strings.LastIndexByte(value, '-')
	if lastDash < 0 {
		return value, 0
	}
	number, _ := strconv.Atoi(value[lastDash+1:])
	return value[:lastDash], number
}

func extensionStringIn(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func buildExtensionTraceabilityMappingSource(document []byte) (map[string]any, error) {
	clauses, err := extractExtensionClauses(document)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(document)
	mappings := make([]any, 0, len(clauses))
	for _, clause := range clauses {
		mappings = append(mappings, map[string]any{
			"source_start_byte":        clause.start,
			"source_end_byte":          clause.end,
			"parent_anchor_kind":       clause.parentAnchorKind,
			"parent_anchor_id":         clause.parentAnchorID,
			"clause_kind":              clause.clauseKind,
			"requirement_ids":          clause.requirementIDs,
			"acceptance_criterion_ids": clause.acceptanceIDs,
			"verification_ids":         clause.verificationIDs,
		})
	}
	return map[string]any{
		"schema_id":                  "cartulary.extension_traceability_mapping_source.v1",
		"extensions_document_sha256": hex.EncodeToString(digest[:]),
		"mappings":                   mappings,
	}, nil
}

func marshalExtensionTraceabilityMappingSource(document []byte) ([]byte, error) {
	source, err := buildExtensionTraceabilityMappingSource(document)
	if err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func refreshExtensionOwnerInputs(root, outputRoot string) error {
	ownersDir := filepath.Join(root, "contracts", "extensions", "owners")
	ownerEntries, err := os.ReadDir(ownersDir)
	if err != nil {
		return err
	}
	manifests := map[string]map[string]any{}
	ownerChanged := map[string]bool{}
	for _, entry := range ownerEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		manifest, err := readExtensionMaintenanceObject(filepath.Join(ownersDir, entry.Name()))
		if err != nil {
			return err
		}
		changed, err := refreshExtensionOwnerManifestDocument(root, manifest)
		if err != nil {
			return fmt.Errorf("refresh %s: %w", entry.Name(), err)
		}
		ownerID := stringValue(manifest["owner_id"])
		manifests[ownerID] = manifest
		ownerChanged[ownerID] = changed
	}

	fragmentsDir := filepath.Join(root, "contracts", "extensions", "fragments")
	fragmentEntries, err := os.ReadDir(fragmentsDir)
	if err != nil {
		return err
	}
	fragmentsByRef := map[string]map[string]any{}
	for _, entry := range fragmentEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		fragment, err := readExtensionMaintenanceObject(filepath.Join(fragmentsDir, entry.Name()))
		if err != nil {
			return err
		}
		manifest := manifests[stringValue(fragment["owner_id"])]
		if manifest == nil {
			return fmt.Errorf("fragment %s has no owner manifest", entry.Name())
		}
		fragmentRef := filepath.ToSlash(filepath.Join("contracts", "extensions", "fragments", entry.Name()))
		fragmentsByRef[fragmentRef] = fragment
		if ownerChanged[stringValue(fragment["owner_id"])] {
			ownerDocument := manifest["owner_document"].(map[string]any)
			fragment["owner_document_sha256"] = ownerDocument["owner_document_sha256"]
			if err := writeExtensionMaintenanceObject(outputRoot, fragmentRef, fragment); err != nil {
				return err
			}
		}
	}

	manifestDigests := map[string]string{}
	for ownerID, manifest := range manifests {
		ownerFragments, _ := objectArray(manifest["owner_fragments"], ownerID+" owner fragments")
		for _, row := range ownerFragments {
			fragmentRef := stringValue(row["owner_fragment_ref"])
			fragment := fragmentsByRef[fragmentRef]
			if fragment == nil {
				return fmt.Errorf("manifest %s refers to missing fragment %s", ownerID, fragmentRef)
			}
			digest, err := extensionCanonicalDigest(fragment)
			if err != nil {
				return err
			}
			row["owner_fragment_sha256"] = digest
		}
		manifestRef := filepath.ToSlash(filepath.Join("contracts", "extensions", "owners", ownerID+".contract-manifest.json"))
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(manifestRef))); err != nil {
			for _, entry := range ownerEntries {
				candidate := filepath.Join(ownersDir, entry.Name())
				decoded, decodeErr := readExtensionMaintenanceObject(candidate)
				if decodeErr == nil && stringValue(decoded["owner_id"]) == ownerID {
					manifestRef = filepath.ToSlash(filepath.Join("contracts", "extensions", "owners", entry.Name()))
					break
				}
			}
		}
		if ownerChanged[ownerID] {
			if err := writeExtensionMaintenanceObject(outputRoot, manifestRef, manifest); err != nil {
				return err
			}
		}
		digest, err := extensionCanonicalDigest(manifest)
		if err != nil {
			return err
		}
		manifestDigests[stringValue(manifest["owner_contract_manifest_id"])] = digest
	}

	dependenciesPath := filepath.Join(root, "contracts", "extensions", "dependencies.json")
	dependencies, err := readExtensionMaintenanceObject(dependenciesPath)
	if err != nil {
		return err
	}
	rows, _ := objectArray(dependencies["dependencies"], "dependencies")
	for _, row := range rows {
		manifest := manifests[stringValue(row["dependency_id"])]
		if manifest == nil {
			return fmt.Errorf("dependency %s has no manifest", row["dependency_id"])
		}
		ownerDocument := manifest["owner_document"].(map[string]any)
		row["owner_document_sha256"] = ownerDocument["owner_document_sha256"]
		manifestID := stringValue(manifest["owner_contract_manifest_id"])
		row["owner_contract_manifest_sha256"] = manifestDigests[manifestID]
	}
	return writeExtensionMaintenanceObject(outputRoot, "contracts/extensions/dependencies.json", dependencies)
}

func refreshExtensionOwnerManifestDocument(root string, manifest map[string]any) (bool, error) {
	ownerDocument := manifest["owner_document"].(map[string]any)
	documentPath := strings.SplitN(stringValue(ownerDocument["owner_document_ref"]), "#", 2)[0]
	document, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(documentPath)))
	if err != nil {
		return false, err
	}
	documentDigest := sha256.Sum256(document)
	actualDigest := hex.EncodeToString(documentDigest[:])
	if ownerDocument["owner_document_sha256"] == actualDigest {
		return false, nil
	}
	ownerDocument["owner_document_sha256"] = actualDigest
	ownerDocument["byte_length"] = len(document)

	anchors, _ := objectArray(manifest["anchors"], "owner anchors")
	oldToNew := map[string]extensionSourceRange{}
	for _, anchor := range anchors {
		kind := stringValue(anchor["anchor_kind"])
		oldKey := fmt.Sprintf("%v:%v", anchor["start_byte"], anchor["end_byte"])
		var current extensionSourceRange
		switch kind {
		case "document":
			current = extensionSourceRange{start: 0, end: len(document)}
		case "req":
			current, err = extensionRequirementSourceRange(document, stringValue(anchor["anchor_id"]))
			if err != nil {
				return false, err
			}
			oldToNew[oldKey] = current
		default:
			continue
		}
		setExtensionAnchorRange(anchor, document, current)
	}
	for _, anchor := range anchors {
		kind := stringValue(anchor["anchor_kind"])
		if kind == "document" || kind == "req" {
			continue
		}
		oldKey := fmt.Sprintf("%v:%v", anchor["start_byte"], anchor["end_byte"])
		current, ok := oldToNew[oldKey]
		if !ok {
			return false, fmt.Errorf("anchor %s has no requirement range owner", anchor["anchor_id"])
		}
		setExtensionAnchorRange(anchor, document, current)
	}
	return true, nil
}

func extensionRequirementSourceRange(document []byte, requirementID string) (extensionSourceRange, error) {
	marker := []byte("**" + requirementID + "**")
	start := bytes.Index(document, marker)
	if start < 0 {
		return extensionSourceRange{}, fmt.Errorf("owner document omits %s", requirementID)
	}
	nextRequirement := regexp.MustCompile(`(?m)^\*\*(?:REQ|NF-REQ)-[A-Za-z0-9.-]+\*\*$`).FindIndex(document[start+len(marker):])
	end := len(document)
	if nextRequirement != nil {
		end = start + len(marker) + nextRequirement[0]
	}
	for end > start && document[end-1] == '\n' {
		end--
	}
	return extensionSourceRange{start: start, end: end}, nil
}

func setExtensionAnchorRange(anchor map[string]any, document []byte, current extensionSourceRange) {
	anchor["start_byte"] = current.start
	anchor["end_byte"] = current.end
	digest := sha256.Sum256(document[current.start:current.end])
	anchor["anchor_sha256"] = hex.EncodeToString(digest[:])
}

func readExtensionMaintenanceObject(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, err := decodeContract(data)
	if err != nil {
		return nil, err
	}
	return asObject(decoded, path)
}

func writeExtensionMaintenanceObject(outputRoot, relativePath string, object map[string]any) error {
	encoded, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(outputRoot, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o600)
}
