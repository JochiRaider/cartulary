package harnessplan

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// TaskSurface is the planning part of the authored task surface: the public
// commands, their input contracts and selection bindings.
type TaskSurface struct {
	Targets     []SurfaceTarget       `json:"targets"`
	MakeRecipes map[string]MakeRecipe `json:"make_recipes"`
}

// SurfaceTarget is one Make target.
type SurfaceTarget struct {
	Name             string            `json:"name"`
	CommandID        string            `json:"command_id"`
	TargetClass      string            `json:"target_class"`
	InputContract    *InputContract    `json:"input_contract"`
	SelectionBinding *SelectionBinding `json:"selection_binding"`
}

// InputContract declares a command's inputs and the constraints between them.
type InputContract struct {
	Inputs      []InputDefinition `json:"inputs"`
	Constraints []InputConstraint `json:"constraints"`
}

// InputDefinition is one declared command input.
type InputDefinition struct {
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	AllowedSources  []string        `json:"allowed_sources"`
	Required        bool            `json:"required"`
	Default         json.RawMessage `json:"default"`
	Values          []string        `json:"values"`
	Min             *int64          `json:"min"`
	Max             *int64          `json:"max"`
	EmptyString     string          `json:"empty_string"`
	Normalization   string          `json:"normalization"`
	SummaryEmission string          `json:"summary_emission"`
	ChildForwarding string          `json:"child_forwarding"`
}

// InputConstraint relates inputs: exactly one of Inputs, or Input requiring
// RequiredInput.
type InputConstraint struct {
	Kind          string   `json:"kind"`
	Inputs        []string `json:"inputs"`
	Input         string   `json:"input"`
	RequiredInput string   `json:"required_input"`
}

// SelectionBinding maps a command's inputs to the work it selects.
type SelectionBinding struct {
	Kind             string            `json:"kind"`
	Input            string            `json:"input"`
	Selection        string            `json:"selection"`
	DefaultSelection string            `json:"default_selection"`
	Variants         map[string]string `json:"variants"`
}

// MakeRecipe is how a public target runs; an operation recipe names the
// operation the target executes.
type MakeRecipe struct {
	Type        string `json:"type"`
	OperationID string `json:"operation_id"`
}

// target returns the named target.
func (s *TaskSurface) target(name string) (SurfaceTarget, bool) {
	index := slices.IndexFunc(s.Targets, func(t SurfaceTarget) bool { return t.Name == name })
	if index < 0 {
		return SurfaceTarget{}, false
	}
	return s.Targets[index], true
}

// projectBrowserLanes applies the topology's lane policy to every
// BROWSER_LANES input: the topology owns its default and maximum.
func (s *TaskSurface) projectBrowserLanes(policy BrowserLanePolicy) {
	for i := range s.Targets {
		contract := s.Targets[i].InputContract
		if contract == nil {
			continue
		}
		for j := range contract.Inputs {
			if contract.Inputs[j].Name == "BROWSER_LANES" {
				maximum := policy.Maximum
				contract.Inputs[j].Default = json.RawMessage(strconv.FormatInt(policy.Default, 10))
				contract.Inputs[j].Max = &maximum
			}
		}
	}
}

// publicInputNames are every input a public command declares. A unit's
// environment clears all of them, so no command inherits another's controls.
func (s *TaskSurface) publicInputNames() []string {
	names := []string{makeInputSources}
	for _, target := range s.Targets {
		if target.TargetClass != "public" || target.InputContract == nil {
			continue
		}
		for _, input := range target.InputContract.Inputs {
			names = append(names, input.Name)
		}
	}
	return dedupe(names)
}

// resolvedInput is one input's value and where it came from.
type resolvedInput struct {
	Value      string
	Default    bool
	Definition InputDefinition
}

// Make origins recorded in CARTULARY_MAKE_INPUT_SOURCES.
const (
	originCommandLine = "cli"
	originEnvironment = "env"
)

var (
	makeSourceToken = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)=(cli|env|file|unset)$`)
	ownerIDPattern  = regexp.MustCompile(`^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$`)
	rowIDPattern    = regexp.MustCompile(`^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}$`)
	digitsPattern   = regexp.MustCompile(`^[0-9]+$`)
)

// resolveInputs resolves a command's declared inputs from the runtime's
// environment and the Make origins it records. Public preflight has already
// admitted the invocation (undeclared inputs, path safety); resolution applies
// sources, defaults, normalization, types and constraints to the values the
// plan consumes and fails closed on anything else.
func resolveInputs(target SurfaceTarget, environment map[string]string) (map[string]resolvedInput, error) {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s: %s", ErrSelection, target.Name, fmt.Sprintf(format, args...))
	}
	origins := map[string]string{}
	for _, token := range strings.Fields(environment[makeInputSources]) {
		match := makeSourceToken.FindStringSubmatch(token)
		if match == nil {
			return nil, fail("invalid %s token %q", makeInputSources, token)
		}
		if _, seen := origins[match[1]]; seen {
			return nil, fail("%s repeats %s", makeInputSources, match[1])
		}
		origins[match[1]] = match[2]
	}
	resolved := map[string]resolvedInput{}
	if target.InputContract == nil {
		return resolved, nil
	}
	for _, input := range target.InputContract.Inputs {
		raw, present := environment[input.Name]
		origin := origins[input.Name]
		useDefault := func() error {
			if input.Required {
				return fail("%s is required", input.Name)
			}
			if value, ok := defaultValue(input.Default); ok {
				resolved[input.Name] = resolvedInput{Value: value, Default: true, Definition: input}
			}
			return nil
		}
		if !present || origin != originCommandLine && origin != originEnvironment {
			if err := useDefault(); err != nil {
				return nil, err
			}
			continue
		}
		if origin == originEnvironment && !slices.Contains(input.AllowedSources, "environment") {
			if input.Required {
				return nil, fail("%s must come from an allowed source", input.Name)
			}
			continue
		}
		value := raw
		switch input.Normalization {
		case "trim", "path_token":
			value = strings.TrimSpace(value)
		case "trim_lowercase":
			value = strings.ToLower(strings.TrimSpace(value))
		}
		if value == "" {
			switch {
			case input.EmptyString == "invalid" || input.Required:
				return nil, fail("%s must not be empty", input.Name)
			case input.EmptyString == "false":
				resolved[input.Name] = resolvedInput{Value: "false", Definition: input}
			case input.EmptyString == "omitted":
				if err := useDefault(); err != nil {
					return nil, err
				}
			}
			continue
		}
		value, err := typedValue(input, value)
		if err != nil {
			return nil, fail("%v", err)
		}
		resolved[input.Name] = resolvedInput{Value: value, Definition: input}
	}
	for _, constraint := range target.InputContract.Constraints {
		present := func(name string) bool { _, ok := resolved[name]; return ok }
		switch constraint.Kind {
		case "exactly_one":
			count := 0
			for _, name := range constraint.Inputs {
				if present(name) {
					count++
				}
			}
			if count != 1 {
				return nil, fail("requires exactly one of %s", strings.Join(constraint.Inputs, ", "))
			}
		case "requires":
			if present(constraint.Input) && !present(constraint.RequiredInput) {
				return nil, fail("%s requires %s", constraint.Input, constraint.RequiredInput)
			}
		default:
			return nil, fail("unknown input constraint %s", constraint.Kind)
		}
	}
	return resolved, nil
}

// defaultValue renders an authored default the way it is declared: strings
// as themselves, booleans and numbers as their JSON text.
func defaultValue(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", false
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text, true
	}
	return string(raw), true
}

// typedValue checks and canonicalizes a value of the input types graph
// commands declare. Any other type is outside the planner's contract.
func typedValue(input InputDefinition, value string) (string, error) {
	switch input.Type {
	case "enum":
		if !slices.Contains(input.Values, value) {
			return "", fmt.Errorf("%s must be one of %s", input.Name, strings.Join(input.Values, ", "))
		}
	case "exact_1_bool":
		if value != "1" {
			return "", fmt.Errorf("%s must be exactly 1 when set", input.Name)
		}
	case "path", "result_selector":
		if strings.ContainsRune(value, 0) {
			return "", fmt.Errorf("%s must not contain NUL", input.Name)
		}
	case "positive_integer":
		minimum, maximum := int64(1), int64(999999999)
		if input.Min != nil {
			minimum = *input.Min
		}
		if input.Max != nil {
			maximum = *input.Max
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if !digitsPattern.MatchString(value) || err != nil || parsed < minimum || parsed > maximum {
			return "", fmt.Errorf("%s must be an integer between %d and %d", input.Name, minimum, maximum)
		}
		return strconv.FormatInt(parsed, 10), nil
	case "owner_id":
		if !ownerIDPattern.MatchString(value) {
			return "", fmt.Errorf("%s must be an owner ID", input.Name)
		}
	case "row_ids":
		tokens := strings.Split(value, ",")
		for i := range tokens {
			tokens[i] = strings.TrimSpace(tokens[i])
			if !rowIDPattern.MatchString(tokens[i]) {
				return "", fmt.Errorf("%s must contain unique row IDs", input.Name)
			}
		}
		if len(value) > 4096 || len(dedupe(tokens)) != len(tokens) {
			return "", fmt.Errorf("%s must contain at most 4096 characters of unique row IDs", input.Name)
		}
		return strings.Join(dedupe(tokens), ","), nil
	default:
		return "", fmt.Errorf("%s has input type %s, which graph commands do not declare", input.Name, input.Type)
	}
	return value, nil
}
