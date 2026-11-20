package openapiassembly

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// errorRegistryPath is the canonical error registry (Core 01 REQ-01-233) that
// declared codes must match.
const errorRegistryPath = "contracts/errors/index.json"

const errorCodesMember = "x-cartulary-error-codes"

// validateDeclaredErrorCodes enforces Core 01 REQ-01-702 on the assembled
// document. Every operation response with status 400 or higher, declared
// directly or through a component response, carries a non-empty, strictly
// ascending x-cartulary-error-codes list, and the registry assigns each code
// exactly that status. No operation may declare a pre-resolution code
// (REQ-01-707). The registry is read only when an error response exists.
// Each operation's security and reserved-family membership then fix its
// implied declarations (validateImpliedErrorCodes).
func validateDeclaredErrorCodes(document *orderedValue, loadRegistry func() (map[string]int, error)) error {
	paths, ok := objectMember(document, "paths")
	if !ok {
		return nil
	}
	var registry map[string]int
	for _, path := range paths.object {
		for _, method := range path.value.object {
			if _, ok := operationKeys[method.name]; !ok {
				continue
			}
			responses, ok := objectMember(method.value, "responses")
			if !ok {
				continue
			}
			for _, response := range responses.object {
				status, err := strconv.Atoi(response.name)
				if err != nil || status < 400 {
					continue
				}
				label := strings.ToUpper(method.name) + " " + path.name + " " + response.name
				declaration := response.value
				if ref, referenced := objectMember(declaration, "$ref"); referenced {
					resolved, found := resolvePointer(document, ref.scalar.(string))
					if !found {
						return fmt.Errorf("%s: unresolved reference %q", label, ref.scalar)
					}
					declaration = resolved
				}
				if registry == nil {
					if registry, err = loadRegistry(); err != nil {
						return err
					}
				}
				if err := validateErrorCodeList(declaration, status, registry); err != nil {
					return fmt.Errorf("%s: %w", label, err)
				}
			}
			if err := validateImpliedErrorCodes(document, method.value, responses); err != nil {
				return fmt.Errorf("%s %s %w", strings.ToUpper(method.name), path.name, err)
			}
		}
	}
	return nil
}

// availabilityMember names the extension profile whose reserved route family
// an operation belongs to; a base operation does not declare it.
const availabilityMember = "x-cartulary-availability"

// impliedDeclarationContext is what an implied declaration depends on: the
// operation's security alternatives and whether it belongs to a route family
// reserved for an extension profile.
type impliedDeclarationContext struct {
	alternatives [][]string
	reserved     bool
}

// impliedErrorCodes are the declarations Core 01 REQ-01-702 implies, each
// enforced in both directions: the response at status declares code exactly
// when applies holds. Declared security implies the first three; every
// operation implies 500 internal_error and 503 service_unavailable; and an
// operation in a reserved extension route family, which is the operation that
// declares x-cartulary-availability, implies 404
// extension_profile_not_claimed. A code declared at another status is a
// registry status mismatch (validateErrorCodeList).
var impliedErrorCodes = []struct {
	status, code, implied, notImplied string
	applies                           func(operation impliedDeclarationContext) bool
}{
	{
		status: "401", code: "session_required",
		implied:    "the operation declares a security requirement",
		notImplied: "the operation declares no security requirement",
		applies: func(operation impliedDeclarationContext) bool {
			return slices.ContainsFunc(operation.alternatives, func(schemes []string) bool { return len(schemes) > 0 })
		},
	},
	{
		status: "403", code: "csrf_verification_failed",
		implied:    "an alternative includes csrfHeader",
		notImplied: "no alternative includes csrfHeader",
		applies: func(operation impliedDeclarationContext) bool {
			return slices.ContainsFunc(operation.alternatives, func(schemes []string) bool { return slices.Contains(schemes, "csrfHeader") })
		},
	},
	{
		status: "409", code: "credential_bootstrap_rejected",
		implied:    "an alternative includes bearerSession or credentialBootstrapBearer",
		notImplied: "no alternative includes bearerSession or credentialBootstrapBearer",
		applies: func(operation impliedDeclarationContext) bool {
			return slices.ContainsFunc(operation.alternatives, func(schemes []string) bool {
				return slices.Contains(schemes, "bearerSession") || slices.Contains(schemes, "credentialBootstrapBearer")
			})
		},
	},
	{
		status: "500", code: "internal_error",
		implied: "every operation may fail internally (REQ-01-229)",
		applies: func(impliedDeclarationContext) bool { return true },
	},
	{
		status: "503", code: "service_unavailable",
		implied: "the admission gate may refuse every operation",
		applies: func(impliedDeclarationContext) bool { return true },
	},
	{
		status: "404", code: "extension_profile_not_claimed",
		implied:    "the operation declares " + availabilityMember + ", so it belongs to a reserved extension route family",
		notImplied: "the operation declares no " + availabilityMember + ", so it belongs to no reserved extension route family",
		applies: func(operation impliedDeclarationContext) bool {
			return operation.reserved
		},
	},
}

// validateImpliedErrorCodes enforces the implied declarations of Core 01
// REQ-01-702 in both directions: an implied code must be declared, directly
// or through a component response, and a declared one must be implied. An
// operation without explicit security declares no requirement, and one
// without x-cartulary-availability belongs to no reserved route family.
func validateImpliedErrorCodes(document, operation, responses *orderedValue) error {
	var context impliedDeclarationContext
	if security, ok := objectMember(operation, "security"); ok {
		if security.kind != arrayKind {
			return errors.New("security must be an array")
		}
		for _, requirement := range security.array {
			if requirement.kind != objectKind {
				return errors.New("security requirement must be an object")
			}
			schemes := make([]string, 0, len(requirement.object))
			for _, scheme := range requirement.object {
				schemes = append(schemes, scheme.name)
			}
			context.alternatives = append(context.alternatives, schemes)
		}
	}
	if availability, ok := objectMember(operation, availabilityMember); ok {
		if availability.kind != stringKind || availability.scalar.(string) == "" {
			return fmt.Errorf("%s must be a non-empty string", availabilityMember)
		}
		context.reserved = true
	}
	for _, rule := range impliedErrorCodes {
		declared := false
		if response, ok := objectMember(responses, rule.status); ok {
			if ref, referenced := objectMember(response, "$ref"); referenced {
				resolved, found := resolvePointer(document, ref.scalar.(string))
				if !found {
					return fmt.Errorf("%s: unresolved reference %q", rule.status, ref.scalar)
				}
				response = resolved
			}
			if codes, ok := objectMember(response, errorCodesMember); ok && codes.kind == arrayKind {
				declared = slices.ContainsFunc(codes.array, func(value *orderedValue) bool {
					return value.kind == stringKind && value.scalar.(string) == rule.code
				})
			}
		}
		switch applies := rule.applies(context); {
		case applies && !declared:
			return fmt.Errorf("%s: %s must be declared because %s", rule.status, rule.code, rule.implied)
		case !applies && declared:
			return fmt.Errorf("%s: %s must not be declared because %s", rule.status, rule.code, rule.notImplied)
		}
	}
	return nil
}

func validateErrorCodeList(response *orderedValue, status int, registry map[string]int) error {
	codes, ok := objectMember(response, errorCodesMember)
	if !ok {
		return fmt.Errorf("error response requires %s", errorCodesMember)
	}
	if codes.kind != arrayKind || len(codes.array) == 0 {
		return fmt.Errorf("%s must be a non-empty array", errorCodesMember)
	}
	previous := ""
	for index, value := range codes.array {
		if value.kind != stringKind {
			return fmt.Errorf("%s[%d] must be a string", errorCodesMember, index)
		}
		code := value.scalar.(string)
		if index > 0 && code <= previous {
			return fmt.Errorf("%s must be ascending without duplicates: %q follows %q", errorCodesMember, code, previous)
		}
		previous = code
		if _, preResolution := preResolutionErrorCodes[code]; preResolution {
			return fmt.Errorf("%s names %q, which belongs to no operation", errorCodesMember, code)
		}
		registered, known := registry[code]
		if !known {
			return fmt.Errorf("%s names unregistered code %q", errorCodesMember, code)
		}
		if registered != status {
			return fmt.Errorf("%s names %q, which the registry assigns status %d", errorCodesMember, code, registered)
		}
	}
	return nil
}

func loadErrorRegistry(repository *os.Root) (map[string]int, error) {
	content, err := repository.ReadFile(errorRegistryPath)
	if err != nil {
		return nil, fmt.Errorf("read error registry: %w", err)
	}
	var document struct {
		Errors []struct {
			Code       string `json:"code"`
			HTTPStatus int    `json:"http_status"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		return nil, fmt.Errorf("decode error registry: %w", err)
	}
	if len(document.Errors) == 0 {
		return nil, errors.New("error registry declares no codes")
	}
	registry := make(map[string]int, len(document.Errors))
	for _, entry := range document.Errors {
		registry[entry.Code] = entry.HTTPStatus
	}
	return registry, nil
}
