package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/google/uuid"
)

// AdmissionMode records how a bound operation is admitted.
type AdmissionMode string

const (
	// AdmissionOwner: the owner's handler authenticates and reads its path
	// itself. Route resolution (REQ-01-707 step 1) still applies.
	AdmissionOwner AdmissionMode = "owner"
	// AdmissionCentral: the registry admits the request in REQ-01-707 order
	// before the handler runs.
	AdmissionCentral AdmissionMode = "central"
)

// SecurityRequirement is the projection of an operation's declared OpenAPI
// security that central admission authenticates by (Core 01 REQ-01-703).
type SecurityRequirement struct {
	// Public: no alternative names a scheme, so the operation never
	// authenticates a session or bearer session.
	Public bool
	// CookieCSRF: the session-cookie alternative pairs CSRF, so a cookie
	// request must pass the double-submit check.
	CookieCSRF bool
	// Bootstrap: an alternative admits a credential bootstrap bearer
	// (credentialBootstrapBearer). The authenticator admits a bootstrap token
	// only for such an operation; elsewhere it is rejected.
	Bootstrap bool
}

// Authenticator establishes the principal of a centrally admitted operation by
// its declared security (REQ-01-707 step 2). The principal's type and accessor
// belong to the implementation.
type Authenticator interface {
	// Authenticate returns the request carrying the principal in its context,
	// or the authentication failure.
	Authenticate(r *http.Request, requirement SecurityRequirement) (*http.Request, *APIError)
	// VerifyCSRF checks the cookie CSRF proof of a request that Authenticate
	// already admitted without it, against the principal Authenticate
	// established and without another session lookup. A bearer session needs
	// no proof. An owner family whose gate authenticates before operation
	// selection (Core 01 §17.4) verifies each selected operation's declared
	// cookie CSRF here.
	VerifyCSRF(r *http.Request) *APIError
}

// OwnerGate is an owner's admission stage inside its own order (Core 01
// REQ-01-707 family clause). Central admission calls it with the operation ID
// after authentication and the path parameters before the gate; it returns the
// request to continue with, or the rejection. Network Flow's gate decides
// incident visibility, role and lifecycle (NF NLSpec Table 21-C level 2).
type OwnerGate func(r *http.Request, operationID string) (*http.Request, *APIError)

func securityRequirementOf(operation ContractOperation) SecurityRequirement {
	declared := newDeclaredSecurity(operation)
	return SecurityRequirement{Public: declared.none, CookieCSRF: declared.cookieCSRF, Bootstrap: declared.bootstrap}
}

// admittedPathParameter is a projected path parameter with its compiled
// pattern.
type admittedPathParameter struct {
	PathParameter
	pattern *regexp.Regexp
}

func compilePathParameters(operation ContractOperation) ([]admittedPathParameter, error) {
	parameters := make([]admittedPathParameter, 0, len(operation.PathParameters))
	for _, parameter := range operation.PathParameters {
		admitted := admittedPathParameter{PathParameter: parameter}
		switch {
		case parameter.Format == "uuid" && parameter.Pattern == "":
		case parameter.Format == "" && parameter.Pattern != "":
			pattern, err := regexp.Compile(parameter.Pattern)
			if err != nil {
				return nil, fmt.Errorf("public operation %q path parameter %q pattern: %w", operation.OperationID, parameter.Name, err)
			}
			admitted.pattern = pattern
		default:
			return nil, fmt.Errorf("public operation %q path parameter %q declares no supported format", operation.OperationID, parameter.Name)
		}
		if parameter.NotFoundCode == "" || parameter.NotFoundStatus < http.StatusBadRequest {
			return nil, fmt.Errorf("public operation %q path parameter %q declares no not-found response", operation.OperationID, parameter.Name)
		}
		parameters = append(parameters, admitted)
	}
	return parameters, nil
}

// admit returns the parameter's typed value: a uuid.UUID for format uuid, the
// decoded string for a pattern. A value outside the declared format is not
// admitted.
func (parameter admittedPathParameter) admit(raw string) (any, bool) {
	if raw == "" || !utf8.ValidString(raw) {
		return nil, false
	}
	length := utf8.RuneCountInString(raw)
	if length < parameter.MinLength || parameter.MaxLength > 0 && length > parameter.MaxLength {
		return nil, false
	}
	if parameter.pattern != nil {
		if !parameter.pattern.MatchString(raw) {
			return nil, false
		}
		return raw, true
	}
	if !isHyphenatedUUID(raw) {
		return nil, false
	}
	value, err := uuid.Parse(raw)
	if err != nil {
		return nil, false
	}
	return value, true
}

// isHyphenatedUUID accepts only the hyphenated 36-character form, in either
// case (REQ-01-707).
func isHyphenatedUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		switch index {
		case 8, 13, 18, 23:
			if character != '-' {
				return false
			}
		default:
			if !('0' <= character && character <= '9' || 'a' <= character && character <= 'f' || 'A' <= character && character <= 'F') {
				return false
			}
		}
	}
	return true
}

type admittedPathValuesContextKey struct{}

// admitOperation authenticates by the declared security, then admits each
// path parameter in template order (REQ-01-707 steps 2 and 3). With an owner
// gate, the parameters before the gate are admitted, then the gate runs, then
// the parameters after it (REQ-01-707 family clause).
func admitOperation(requirement SecurityRequirement, authenticator Authenticator, operationID string, parameters []admittedPathParameter, gate OwnerGate, handler http.HandlerFunc) http.HandlerFunc {
	before, after := splitAtOwnerGate(parameters)
	return func(w http.ResponseWriter, r *http.Request) {
		if !requirement.Public {
			authenticated, apiErr := authenticator.Authenticate(r, requirement)
			if apiErr != nil {
				WriteAPIError(w, r, apiErr)
				return
			}
			r = authenticated
		}
		values := make(map[string]any, len(parameters))
		if !admitParameterValues(w, r, before, values) {
			return
		}
		if gate != nil {
			gated, apiErr := gate(withAdmittedPathValues(r, values), operationID)
			if apiErr != nil {
				WriteAPIError(w, r, apiErr)
				return
			}
			r = gated
		}
		if !admitParameterValues(w, r, after, values) {
			return
		}
		handler(w, withAdmittedPathValues(r, values))
	}
}

// splitAtOwnerGate separates the parameters admitted before an owner gate from
// those admitted after it, each in template order.
func splitAtOwnerGate(parameters []admittedPathParameter) (before, after []admittedPathParameter) {
	for _, parameter := range parameters {
		if parameter.AfterOwnerGate {
			after = append(after, parameter)
		} else {
			before = append(before, parameter)
		}
	}
	return before, after
}

// admitPathParameters admits each path parameter in template order before the
// handler runs.
func admitPathParameters(parameters []admittedPathParameter, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		values := make(map[string]any, len(parameters))
		if !admitParameterValues(w, r, parameters, values) {
			return
		}
		handler(w, withAdmittedPathValues(r, values))
	}
}

// admitParameterValues admits each parameter in order into values. The first
// value outside its declared format receives the parameter's declared
// not-found response, identical to the response for a missing resource, and
// the request ends.
func admitParameterValues(w http.ResponseWriter, r *http.Request, parameters []admittedPathParameter, values map[string]any) bool {
	for _, parameter := range parameters {
		value, ok := parameter.admit(r.PathValue(parameter.Name))
		if !ok {
			_ = WriteErrorWithOptions(w, r, parameter.NotFoundStatus, parameter.NotFoundCode, parameter.NotFoundCode, cloneNotFoundDetails(parameter.NotFoundDetails), ErrorOptions{})
			return false
		}
		values[parameter.Name] = value
	}
	return true
}

// withAdmittedPathValues carries a copy of the admitted values, so values
// admitted later never change what an earlier stage observed.
func withAdmittedPathValues(r *http.Request, values map[string]any) *http.Request {
	admitted := make(map[string]any, len(values))
	for name, value := range values {
		admitted[name] = value
	}
	return r.WithContext(context.WithValue(r.Context(), admittedPathValuesContextKey{}, admitted))
}

// ParseUUIDPathValue parses a path value in REQ-01-707's uuid form: only the
// hyphenated 36-character form, in either case. It serves a path that central
// admission does not admit, the WebSocket upgrade (Core 01 REQ-01-708).
func ParseUUIDPathValue(raw string) (uuid.UUID, bool) {
	if !isHyphenatedUUID(raw) {
		return uuid.UUID{}, false
	}
	value, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, false
	}
	return value, true
}

// PathUUID returns a centrally admitted uuid path parameter. It reports false
// when the operation was not admitted centrally or declares no such uuid
// parameter, which is a binding error.
func PathUUID(r *http.Request, name string) (uuid.UUID, bool) {
	values, _ := r.Context().Value(admittedPathValuesContextKey{}).(map[string]any)
	value, ok := values[name].(uuid.UUID)
	return value, ok
}

// PathString returns a centrally admitted pattern path parameter. It reports
// false when the operation was not admitted centrally or declares no such
// pattern parameter, which is a binding error.
func PathString(r *http.Request, name string) (string, bool) {
	values, _ := r.Context().Value(admittedPathValuesContextKey{}).(map[string]any)
	value, ok := values[name].(string)
	return value, ok
}
