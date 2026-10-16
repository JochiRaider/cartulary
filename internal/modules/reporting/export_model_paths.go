package reporting

import "strings"

// Exact model paths are addresses, not a query language. Numeric indexes are
// diagnostic-only; wildcards belong to the separate template binding contract.
type modelPathStep struct {
	Value    string
	Key      bool
	Wildcard bool
}

func validExactExportModelPath(path string) bool {
	_, valid := parseExactModelPath(path)
	return valid
}

func parseExactModelPath(path string) ([]modelPathStep, bool) {
	return parseModelPath(path, false)
}

func parseModelPath(path string, allowWildcard bool) ([]modelPathStep, bool) {
	steps := []modelPathStep{}
	wildcards := 0
	if path == "" || path[0] != '$' {
		return nil, false
	}
	for index := 1; index < len(path); {
		switch path[index] {
		case '.':
			index++
			start := index
			for index < len(path) && path[index] != '.' && path[index] != '[' {
				index++
			}
			if !reportingMemberName.MatchString(path[start:index]) {
				return nil, false
			}
			steps = append(steps, modelPathStep{Value: path[start:index]})
		case '[':
			index++
			if allowWildcard && index+1 < len(path) && path[index:index+2] == "*]" {
				wildcards++
				if wildcards > 1 {
					return nil, false
				}
				steps = append(steps, modelPathStep{Key: true, Wildcard: true})
				index += 2
				continue
			}
			if index >= len(path) || path[index] != '\'' {
				return nil, false
			}
			index++
			var id strings.Builder
			closed := false
			for index < len(path) {
				c := path[index]
				index++
				if c == '\'' {
					if index >= len(path) || path[index] != ']' {
						return nil, false
					}
					index++
					closed = true
					break
				}
				if c == '\\' {
					if index >= len(path) || path[index] != '\\' && path[index] != '\'' {
						return nil, false
					}
					c = path[index]
					index++
				}
				id.WriteByte(c)
			}
			if !closed || !validReportingIdentifier(id.String()) {
				return nil, false
			}
			steps = append(steps, modelPathStep{Value: id.String(), Key: true})
		default:
			return nil, false
		}
	}
	return steps, true
}

func keyedModelPath(parent, member, id string) (string, error) {
	if !reportingMemberName.MatchString(member) || !validReportingIdentifier(id) {
		return "", newRenderValidationError("export_model_invalid", "invalid_scalar_value")
	}
	escaped := strings.ReplaceAll(strings.ReplaceAll(id, "\\", "\\\\"), "'", "\\'")
	return parent + "." + member + "['" + escaped + "']", nil
}

func formatExactModelPath(steps []modelPathStep) string {
	var result strings.Builder
	result.WriteByte('$')
	for _, step := range steps {
		if step.Key {
			result.WriteString("['")
			result.WriteString(strings.ReplaceAll(strings.ReplaceAll(step.Value, "\\", "\\\\"), "'", "\\'"))
			result.WriteString("']")
		} else {
			result.WriteByte('.')
			result.WriteString(step.Value)
		}
	}
	return result.String()
}

// Each selected value has two separate identities: its source binding and its
// address in the admitted model. Resolve exact rules only against the latter;
// source paths continue to identify provenance and token subjects.
func redactionFieldModelPaths(model SnapshotContent) (map[string]string, error) {
	paths := map[string]string{}
	addresses := map[string]bool{}
	var visit func(string, string, []ReportingBlock) error
	visit = func(parent, member string, blocks []ReportingBlock) error {
		for _, block := range blocks {
			blockPath, err := keyedModelPath(parent, member, block.BlockID)
			if err != nil {
				return err
			}
			for _, field := range block.Fields {
				fieldPath, err := keyedModelPath(blockPath, "fields", field.FieldKey)
				if err != nil {
					return err
				}
				if len(field.SourceRefs) == 0 || field.SourceRefs[0] == "" {
					return newRenderValidationError("export_model_invalid", "dangling_source_ref")
				}
				source := field.SourceRefs[0]
				if paths[source] != "" || addresses[fieldPath] {
					return newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
				}
				paths[source], addresses[fieldPath] = fieldPath, true
			}
			if err := visit(blockPath, "children", block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	for _, section := range model.Sections {
		sectionPath, err := keyedModelPath("$", "sections", section.SectionID)
		if err != nil {
			return nil, err
		}
		if err := visit(sectionPath, "blocks", section.Blocks); err != nil {
			return nil, err
		}
	}
	return paths, nil
}
