package packformat

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
)

var diagnosticVocabulary = loadDiagnosticVocabulary()

type diagnosticWords struct{ fields, limits, tokens, paths map[string]bool }

func loadDiagnosticVocabulary() diagnosticWords {
	v := diagnosticWords{map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}}
	var catalog struct {
		Fields []string `json:"virtual_members"`
		Tokens []string `json:"tokens"`
		Paths  []string `json:"fixed_paths"`
	}
	if err := json.Unmarshal(projection("diagnostic_vocabulary.v1.json"), &catalog); err != nil {
		panic("invalid diagnostic vocabulary")
	}
	for _, field := range catalog.Fields {
		v.fields[field] = true
	}
	for _, token := range catalog.Tokens {
		v.tokens[token] = true
	}
	for _, path := range catalog.Paths {
		v.paths[path] = true
	}
	// Only names declared by compiled owner schemas may appear in a diagnostic.
	// Signed extensions and hostile unknown input names never enlarge this set.
	var walk func(any)
	walk = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			if properties, ok := node["properties"].(map[string]any); ok {
				for name := range properties {
					v.fields[name] = true
				}
			}
			for _, child := range node {
				walk(child)
			}
		case []any:
			for _, child := range node {
				walk(child)
			}
		}
	}
	for _, artifact := range contractreferencepacks.Artifacts {
		if !strings.HasSuffix(artifact.Path, ".schema.json") {
			continue
		}
		var value any
		if err := json.Unmarshal([]byte(artifact.JSON), &value); err != nil {
			panic("invalid diagnostic schema projection")
		}
		walk(value)
	}
	var limits map[string]any
	if err := json.Unmarshal(projection("limits.v1.json"), &limits); err != nil {
		panic("invalid diagnostic limit projection")
	}
	for _, group := range []string{"fixed", "defaults", "ranges"} {
		for name := range limits[group].(map[string]any) {
			v.limits[name] = true
		}
	}
	return v
}

func diagnosticDecimal(value string) bool {
	if value == "" || len(value) > 16 || (len(value) > 1 && value[0] == '0') {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	v, err := strconv.ParseUint(value, 10, 64)
	return err == nil && v <= 9007199254740991
}

// Measurements are decimal strings, so unsigned 64-bit guard counts do not
// inherit JSON's exact-integer ceiling. Paths retain their separate safe bound.
func diagnosticMeasurement(value string) bool {
	if value == "" || len(value) > 20 || len(value) > 1 && value[0] == '0' {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	_, err := strconv.ParseUint(value, 10, 64)
	return err == nil
}

func safeDiagnosticPath(path string) bool {
	if len(path) == 0 || len(path) > 4096 {
		return false
	}
	if diagnosticVocabulary.paths[path] {
		return true
	}
	if strings.HasPrefix(path, "metadata/") && strings.HasSuffix(path, ".root.json") {
		version := strings.TrimSuffix(strings.TrimPrefix(path, "metadata/"), ".root.json")
		return version != "0" && diagnosticDecimal(version)
	}
	if path[0] != '$' {
		return false
	}
	for rest := path[1:]; rest != ""; {
		if strings.HasPrefix(rest, `["<unknown>"]`) {
			rest = rest[len(`["<unknown>"]`):]
			continue
		}
		switch rest[0] {
		case '.':
			end := 1
			for end < len(rest) && rest[end] != '.' && rest[end] != '[' {
				end++
			}
			if !diagnosticVocabulary.fields[rest[1:end]] {
				return false
			}
			rest = rest[end:]
		case '[':
			end := strings.IndexByte(rest, ']')
			if end < 2 {
				return false
			}
			segment := rest[1:end]
			if !diagnosticDecimal(segment) {
				var name string
				if err := json.Unmarshal([]byte(segment), &name); err != nil || !diagnosticVocabulary.fields[name] {
					return false
				}
				encoded, _ := json.Marshal(name)
				if string(encoded) != segment {
					return false
				}
			}
			rest = rest[end+1:]
		default:
			return false
		}
	}
	return true
}

func validateFinding(f Finding) error {
	if !safeDiagnosticPath(f.Path) {
		return errors.New("reference pack: unsafe diagnostic path")
	}
	if f.EntryID != nil && (len(*f.EntryID) > 512 || !singleLine(*f.EntryID, 512)) {
		return errors.New("reference pack: unsafe diagnostic identity")
	}
	if f.Details.LimitID != nil && !diagnosticVocabulary.limits[*f.Details.LimitID] {
		return errors.New("reference pack: unknown diagnostic limit")
	}
	for _, token := range []*string{f.Details.ExpectedToken, f.Details.ActualToken} {
		if token != nil && !diagnosticVocabulary.tokens[*token] && !diagnosticMeasurement(*token) {
			return errors.New("reference pack: unknown diagnostic token")
		}
	}
	if key := f.Details.RelatedPackKey; key != nil && (len(*key) > 128 || !keyPattern.MatchString(*key)) {
		return errors.New("reference pack: unsafe related pack key")
	}
	if version := f.Details.RelatedPackVersion; version != nil && !versionPattern.MatchString(*version) {
		return errors.New("reference pack: unsafe related pack version")
	}
	return nil
}
