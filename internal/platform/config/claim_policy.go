package config

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var claimRegistrationIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

type claimCatalog struct {
	registrations []ClaimRegistration
	byPath        map[string]ClaimRegistration
}

func newClaimCatalog(policy ExtensionPolicy) (claimCatalog, error) {
	catalog := claimCatalog{byPath: map[string]ClaimRegistration{}}
	if policy == nil {
		return catalog, nil
	}
	registrations := policy.ClaimRegistrations()
	seenIDs := make(map[string]struct{}, len(registrations))
	previousID := ""
	for _, registration := range registrations {
		if !claimRegistrationIDPattern.MatchString(registration.ID) {
			return claimCatalog{}, fmt.Errorf("extension claim registration id %q is invalid", registration.ID)
		}
		if registration.Path != registration.ID+".claimed" {
			return claimCatalog{}, fmt.Errorf("extension claim registration %q does not own canonical path %q", registration.ID, registration.Path)
		}
		if previousID != "" && registration.ID <= previousID {
			return claimCatalog{}, fmt.Errorf("extension claim registrations are not in canonical id order")
		}
		if _, duplicate := seenIDs[registration.ID]; duplicate {
			return claimCatalog{}, fmt.Errorf("duplicate extension claim registration id %q", registration.ID)
		}
		if _, duplicate := catalog.byPath[registration.Path]; duplicate {
			return claimCatalog{}, fmt.Errorf("duplicate extension claim registration path %q", registration.Path)
		}
		seenIDs[registration.ID] = struct{}{}
		catalog.byPath[registration.Path] = registration
		catalog.registrations = append(catalog.registrations, registration)
		previousID = registration.ID
	}
	return catalog, nil
}

func (catalog claimCatalog) collect(cfg *document, raw map[string]any) []Diagnostic {
	cfg.claims = make(map[string]registeredClaim, len(catalog.registrations))
	diagnostics := make([]Diagnostic, 0)
	for _, registration := range catalog.registrations {
		claim := registeredClaim{id: registration.ID}
		if value, present := rawValueAtPath(raw, registration.Path); present {
			boolean, valid := value.(bool)
			if !valid {
				diagnostics = append(diagnostics, Diagnostic{
					Path:       registration.Path,
					ReasonCode: "type_mismatch",
					Message:    "extension claim value must be Boolean",
				})
			} else {
				claim.value = boolean
			}
		}
		cfg.claims[registration.Path] = claim
	}
	return diagnostics
}

func (catalog claimCatalog) applyOverlay(cfg *document, segments []string, raw string) (bool, *Diagnostic) {
	path := strings.Join(segments, ".")
	registration, present := catalog.byPath[path]
	if !present {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return true, &Diagnostic{
			Path:       path,
			ReasonCode: "type_mismatch",
			Message:    fmt.Sprintf("parse boolean overlay: %v", err),
		}
	}
	if cfg.claims == nil {
		cfg.claims = make(map[string]registeredClaim, len(catalog.registrations))
	}
	cfg.claims[path] = registeredClaim{id: registration.ID, value: value}
	return true, nil
}

func (catalog claimCatalog) ownsPath(path string) bool {
	for _, registration := range catalog.registrations {
		if path == registration.Path || strings.HasPrefix(registration.Path, path+".") {
			return true
		}
	}
	return false
}

func requestedClaimRegistrationIDs(cfg document) []string {
	requested := make([]string, 0)
	for _, claim := range cfg.claims {
		if claim.value {
			requested = append(requested, claim.id)
		}
	}
	sort.Strings(requested)
	return requested
}
