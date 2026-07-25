package enterpriseauth

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/url"
	pathpkg "path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
)

const (
	providerManifestSchemaID = "cartulary.enterprise_auth_providers.v1"
	providerManifestRootPath = "enterprise_authentication.provider_manifest"
)

var (
	providerKeyPattern     = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	secretRefNamePattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,63}$`)
	oidcScopeTokenPattern  = regexp.MustCompile(`^[\x21\x23-\x5b\x5d-\x7e]+$`)
	errDuplicateJSONMember = errors.New("duplicate JSON object member")
)

type DocumentReadFailure string

const (
	DocumentUnavailable DocumentReadFailure = "unavailable"
	DocumentUnsafe      DocumentReadFailure = "unsafe_object"
	DocumentTooLarge    DocumentReadFailure = "too_large"
)

// DocumentReader is the owner port supplied by application assembly. It must
// return immutable bytes from one no-follow, bounded read and never expose the
// underlying host path through its failure.
type DocumentReader interface {
	ReadDocument(absolutePath string, maximumBytes int64) ([]byte, DocumentReadFailure)
}

type DocumentReadFunc func(absolutePath string, maximumBytes int64) ([]byte, DocumentReadFailure)

func (read DocumentReadFunc) ReadDocument(absolutePath string, maximumBytes int64) ([]byte, DocumentReadFailure) {
	return read(absolutePath, maximumBytes)
}

type providerManifestRoot struct {
	ProviderManifestSchemaID string
	Providers                []json.RawMessage
}

type secretRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type OIDCManifestProvider struct {
	ProviderKey           string
	ProviderType          string
	DisplayName           string
	Enabled               bool
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	JWKSURI               string
	ClientID              string
	ClientSecretRef       secretRef
	AdditionalScopes      []string
}

type SAMLManifestProvider struct {
	ProviderKey                string
	ProviderType               string
	DisplayName                string
	Enabled                    bool
	IDPEntityID                string
	SSOURL                     string
	IDPSigningCertificatePaths []string
	SPEntityID                 string
	SubjectSource              authn.EnterpriseAuthSAMLSubjectSource
}

type ProviderReconciliationStore interface {
	ReconcileEnterpriseAuthProviders(context.Context, []authn.EnterpriseAuthProviderDefinition, time.Time) error
}

func ReconcileProviderDefinitions(ctx context.Context, definitions []authn.EnterpriseAuthProviderDefinition, store ProviderReconciliationStore, now time.Time) error {
	if err := store.ReconcileEnterpriseAuthProviders(ctx, definitions, now); err != nil {
		if errors.Is(err, authn.ErrAuthProviderTypeChangeNotSupported) {
			return providerConfigError(providerManifestPathKey, "provider_type_change_not_supported", "enterprise provider type changes require an explicit owner migration")
		}
		return providerConfigError(providerManifestPathKey, "provider_manifest_persist_failed", fmt.Sprintf("reconcile enterprise provider manifest: %v", err))
	}
	return nil
}

func RegisterProviderSecretPurposes(definitions []authn.EnterpriseAuthProviderDefinition, env map[string]string, registry *secretpurpose.Registry) error {
	for index, definition := range definitions {
		if definition.ClientSecretRefName == nil {
			continue
		}
		value, ok := lookupEnv(env, secretRefEnvName(*definition.ClientSecretRefName))
		if !ok || !isValidResolvedSecret(value) {
			return providerConfigError(providerPath(index, "client_secret_ref"), "provider_manifest_secret_missing", "secret_ref_v1 could not be resolved to a safe value")
		}
		if err := registry.Register(*definition.ClientSecretRefName, "enterprise_authentication.oidc."+definition.ProviderKey, []byte(value)); err != nil {
			return providerConfigError(providerPath(index, "client_secret_ref"), "provider_manifest_secret_ref_invalid", "secret reference or resolved material is already registered for another startup purpose")
		}
	}
	return nil
}

func LoadProviderManifest(configuration Configuration, env map[string]string, reader DocumentReader) ([]authn.EnterpriseAuthProviderDefinition, error) {
	normalized, findings := NormalizeAndValidateConfiguration(configuration)
	if len(findings) > 0 {
		finding := findings[0]
		return nil, providerConfigError(finding.Path, finding.ReasonCode, finding.Message)
	}
	if !normalized.Claimed {
		return nil, nil
	}
	if reader == nil {
		return nil, providerConfigError(providerManifestPathKey, "provider_manifest_not_readable", "enterprise provider manifest is unavailable")
	}
	raw, failure := reader.ReadDocument(normalized.ProviderManifestPath, ProviderManifestMaximumSize)
	switch failure {
	case "":
	case DocumentTooLarge:
		return nil, providerConfigError(providerManifestPathKey, "provider_manifest_schema_invalid", "enterprise provider manifest must not exceed 1048576 bytes")
	case DocumentUnsafe:
		return nil, providerConfigError(providerManifestPathKey, "provider_manifest_not_regular_file", "enterprise provider manifest path must reference one regular file")
	default:
		return nil, providerConfigError(providerManifestPathKey, "provider_manifest_not_readable", "enterprise provider manifest is unavailable")
	}
	if !utf8.Valid(raw) {
		return nil, providerConfigError(providerManifestPathKey, "provider_manifest_encoding_invalid", "enterprise provider manifest must be UTF-8 JSON")
	}
	if !json.Valid(raw) {
		_, err := parseProviderManifestRoot(raw)
		return nil, err
	}
	if err := rejectDuplicateObjectMembers(raw); err != nil {
		return nil, providerConfigError(providerManifestRootPath, "provider_manifest_schema_invalid", err.Error())
	}

	root, err := parseProviderManifestRoot(raw)
	if err != nil {
		return nil, err
	}
	if root.ProviderManifestSchemaID != providerManifestSchemaID {
		return nil, providerConfigError(providerManifestRootPath+".provider_manifest_schema_id", "provider_manifest_schema_invalid", fmt.Sprintf("provider_manifest_schema_id must equal %q", providerManifestSchemaID))
	}
	if len(root.Providers) > 32 {
		return nil, providerConfigError(providerManifestRootPath+".providers", "provider_manifest_schema_invalid", "providers must contain at most 32 entries")
	}

	definitions := make([]authn.EnterpriseAuthProviderDefinition, 0, len(root.Providers))
	seenProviderKeys := make(map[string]struct{}, len(root.Providers))
	for index, rawProvider := range root.Providers {
		definition, err := parseProviderDefinition(index, rawProvider, env, reader)
		if err != nil {
			return nil, err
		}
		if _, exists := seenProviderKeys[definition.ProviderKey]; exists {
			return nil, providerConfigError(providerPath(index, "provider_key"), "provider_manifest_schema_invalid", "provider_key values must be unique within the provider manifest")
		}
		seenProviderKeys[definition.ProviderKey] = struct{}{}
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].ProviderKey < definitions[j].ProviderKey
	})
	return definitions, nil
}

func parseProviderManifestRoot(raw []byte) (providerManifestRoot, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return providerManifestRoot{}, providerConfigError(providerManifestPathKey, "provider_manifest_parse_error", fmt.Sprintf("parse enterprise provider manifest JSON: %v", err))
	}
	if root == nil {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath, "provider_manifest_schema_invalid", "enterprise provider manifest must be one JSON object")
	}
	for key := range root {
		switch key {
		case "provider_manifest_schema_id", "providers":
		default:
			return providerManifestRoot{}, providerConfigError(providerManifestRootPath+"."+key, "provider_manifest_schema_invalid", "unknown enterprise provider manifest member")
		}
	}
	schemaIDRaw, ok := root["provider_manifest_schema_id"]
	if !ok || isJSONNull(schemaIDRaw) {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath+".provider_manifest_schema_id", "provider_manifest_schema_invalid", "provider_manifest_schema_id is required")
	}
	var schemaID string
	if err := json.Unmarshal(schemaIDRaw, &schemaID); err != nil || schemaID == "" {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath+".provider_manifest_schema_id", "provider_manifest_schema_invalid", "provider_manifest_schema_id must be a non-null string")
	}

	providersRaw, ok := root["providers"]
	if !ok || isJSONNull(providersRaw) {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath+".providers", "provider_manifest_schema_invalid", "providers is required")
	}
	var providers []json.RawMessage
	if err := json.Unmarshal(providersRaw, &providers); err != nil {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath+".providers", "provider_manifest_schema_invalid", "providers must be an array")
	}
	if providers == nil {
		return providerManifestRoot{}, providerConfigError(providerManifestRootPath+".providers", "provider_manifest_schema_invalid", "providers must be an array")
	}
	return providerManifestRoot{ProviderManifestSchemaID: schemaID, Providers: providers}, nil
}

func parseProviderDefinition(index int, raw json.RawMessage, env map[string]string, reader DocumentReader) (authn.EnterpriseAuthProviderDefinition, error) {
	providerObject, err := decodeProviderObject(index, raw)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	providerType, err := requiredStringMember(providerObject, providerPath(index, "provider_type"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	switch providerType {
	case "oidc":
		return parseOIDCProviderDefinition(index, providerObject, env)
	case "saml":
		return parseSAMLProviderDefinition(index, providerObject, reader)
	default:
		return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "provider_type"), "provider_manifest_schema_invalid", "provider_type must be oidc or saml")
	}
}

func parseOIDCProviderDefinition(index int, provider map[string]json.RawMessage, env map[string]string) (authn.EnterpriseAuthProviderDefinition, error) {
	allowed := map[string]struct{}{
		"provider_key":           {},
		"provider_type":          {},
		"display_name":           {},
		"enabled":                {},
		"issuer":                 {},
		"authorization_endpoint": {},
		"token_endpoint":         {},
		"jwks_uri":               {},
		"client_id":              {},
		"client_secret_ref":      {},
		"additional_scopes":      {},
	}
	if err := rejectUnknownProviderMembers(index, provider, allowed); err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	base, err := parseCommonProvider(index, provider)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	issuer, err := requiredHTTPSURLMember(provider, providerPath(index, "issuer"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	authURL, err := requiredHTTPSURLMember(provider, providerPath(index, "authorization_endpoint"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	tokenURL, err := requiredHTTPSURLMember(provider, providerPath(index, "token_endpoint"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	jwksURI, err := requiredHTTPSURLMember(provider, providerPath(index, "jwks_uri"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	clientID, err := requiredClientIDMember(provider, providerPath(index, "client_id"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	clientSecretRef, err := requiredSecretRef(provider, providerPath(index, "client_secret_ref"), env)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	additionalScopes, err := optionalAdditionalScopes(provider, index)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	base.AuthorizationEndpoint = stringPtr(authURL)
	base.Issuer = stringPtr(issuer)
	base.Audience = stringPtr(clientID)
	base.TokenEndpoint = stringPtr(tokenURL)
	base.JWKSURI = stringPtr(jwksURI)
	base.ClientID = stringPtr(clientID)
	base.ClientSecretRefKind = stringPtr(clientSecretRef.Kind)
	base.ClientSecretRefName = stringPtr(clientSecretRef.Name)
	base.AdditionalScopes = additionalScopes
	return base, nil
}

func parseSAMLProviderDefinition(index int, provider map[string]json.RawMessage, reader DocumentReader) (authn.EnterpriseAuthProviderDefinition, error) {
	allowed := map[string]struct{}{
		"provider_key":                  {},
		"provider_type":                 {},
		"display_name":                  {},
		"enabled":                       {},
		"idp_entity_id":                 {},
		"sso_url":                       {},
		"idp_signing_certificate_paths": {},
		"sp_entity_id":                  {},
		"subject_source":                {},
	}
	if err := rejectUnknownProviderMembers(index, provider, allowed); err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	base, err := parseCommonProvider(index, provider)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	idpEntityID, err := requiredStringMember(provider, providerPath(index, "idp_entity_id"))
	if err != nil || strings.TrimSpace(idpEntityID) == "" {
		return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "idp_entity_id"), "provider_manifest_schema_invalid", "idp_entity_id must be a non-empty string")
	}
	ssoURL, err := requiredHTTPSURLMember(provider, providerPath(index, "sso_url"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	certificates, err := requiredSigningCertificates(provider, index, reader)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	spEntityID, err := requiredStringMember(provider, providerPath(index, "sp_entity_id"))
	if err != nil || strings.TrimSpace(spEntityID) == "" {
		return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "sp_entity_id"), "provider_manifest_schema_invalid", "sp_entity_id must be a non-empty string")
	}
	subjectSource, err := requiredSAMLSubjectSource(provider, index)
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	base.AuthorizationEndpoint = stringPtr(ssoURL)
	base.Issuer = stringPtr(idpEntityID)
	base.Audience = stringPtr(spEntityID)
	base.SAMLIDPEntityID = stringPtr(idpEntityID)
	base.SAMLSSOURL = stringPtr(ssoURL)
	base.SAMLIDPSigningCertificate = certificates
	base.SAMLSPHostEntityID = stringPtr(spEntityID)
	base.SAMLSubjectSource = subjectSource
	return base, nil
}

func parseCommonProvider(index int, provider map[string]json.RawMessage) (authn.EnterpriseAuthProviderDefinition, error) {
	providerKey, err := requiredStringMember(provider, providerPath(index, "provider_key"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	if !providerKeyPattern.MatchString(providerKey) {
		return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "provider_key"), "provider_manifest_schema_invalid", "provider_key is outside the enterprise provider token contract")
	}
	providerType, err := requiredStringMember(provider, providerPath(index, "provider_type"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	displayNameRaw, err := requiredStringMember(provider, providerPath(index, "display_name"))
	if err != nil {
		return authn.EnterpriseAuthProviderDefinition{}, err
	}
	displayName, ok := authn.NormalizeDisplayNameLine(displayNameRaw)
	if !ok {
		return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "display_name"), "provider_manifest_schema_invalid", "display_name must satisfy display_name_line_v1")
	}
	enabled := true
	if raw, ok := provider["enabled"]; ok {
		if isJSONNull(raw) {
			return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "enabled"), "provider_manifest_schema_invalid", "enabled must not be null")
		}
		if err := json.Unmarshal(raw, &enabled); err != nil {
			return authn.EnterpriseAuthProviderDefinition{}, providerConfigError(providerPath(index, "enabled"), "provider_manifest_schema_invalid", "enabled must be a boolean")
		}
	}
	return authn.EnterpriseAuthProviderDefinition{
		ProviderKey:  providerKey,
		ProviderType: providerType,
		DisplayName:  displayName,
		Enabled:      enabled,
	}, nil
}

func decodeProviderObject(index int, raw json.RawMessage) (map[string]json.RawMessage, error) {
	if isJSONNull(raw) {
		return nil, providerConfigError(providerBasePath(index), "provider_manifest_schema_invalid", "provider entry must be an object")
	}
	var provider map[string]json.RawMessage
	if err := json.Unmarshal(raw, &provider); err != nil || provider == nil {
		return nil, providerConfigError(providerBasePath(index), "provider_manifest_schema_invalid", "provider entry must be an object")
	}
	return provider, nil
}

func rejectUnknownProviderMembers(index int, provider map[string]json.RawMessage, allowed map[string]struct{}) error {
	for key := range provider {
		if _, ok := allowed[key]; ok {
			continue
		}
		return providerConfigError(providerPath(index, key), "provider_manifest_schema_invalid", "unknown provider member")
	}
	return nil
}

func requiredStringMember(object map[string]json.RawMessage, path string) (string, error) {
	field := path[strings.LastIndex(path, ".")+1:]
	raw, ok := object[field]
	if !ok || isJSONNull(raw) {
		return "", providerConfigError(path, "provider_manifest_schema_invalid", field+" is required")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", providerConfigError(path, "provider_manifest_schema_invalid", field+" must be a non-null string")
	}
	return value, nil
}

func requiredHTTPSURLMember(object map[string]json.RawMessage, path string) (string, error) {
	value, err := requiredStringMember(object, path)
	if err != nil {
		return "", err
	}
	if !validHTTPSProviderURL(value) {
		return "", providerConfigError(path, "provider_manifest_schema_invalid", "provider URL must be absolute HTTPS and must not include userinfo, query, or fragment")
	}
	return value, nil
}

func requiredClientIDMember(object map[string]json.RawMessage, path string) (string, error) {
	value, err := requiredStringMember(object, path)
	if err != nil {
		return "", err
	}
	if value == "" || utf8.RuneCountInString(value) > 256 || hasControlRunes(value) {
		return "", providerConfigError(path, "provider_manifest_schema_invalid", "client_id must be a non-empty string without controls and at most 256 Unicode scalar values")
	}
	return value, nil
}

func requiredSecretRef(object map[string]json.RawMessage, path string, env map[string]string) (secretRef, error) {
	raw, ok := object["client_secret_ref"]
	if !ok || isJSONNull(raw) {
		return secretRef{}, providerConfigError(path, "provider_manifest_secret_ref_invalid", "client_secret_ref is required")
	}
	var refMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &refMap); err != nil || refMap == nil {
		return secretRef{}, providerConfigError(path, "provider_manifest_secret_ref_invalid", "client_secret_ref must be a secret_ref_v1 object")
	}
	for key := range refMap {
		if key != "kind" && key != "name" {
			return secretRef{}, providerConfigError(path+"."+key, "provider_manifest_secret_ref_invalid", "unknown secret_ref_v1 member")
		}
	}
	kind, err := requiredStringMember(refMap, path+".kind")
	if err != nil {
		return secretRef{}, providerConfigError(path+".kind", "provider_manifest_secret_ref_invalid", "secret_ref_v1 kind is required")
	}
	name, err := requiredStringMember(refMap, path+".name")
	if err != nil {
		return secretRef{}, providerConfigError(path+".name", "provider_manifest_secret_ref_invalid", "secret_ref_v1 name is required")
	}
	if kind != "env" || !isValidSecretRefName(name) {
		return secretRef{}, providerConfigError(path, "provider_manifest_secret_ref_invalid", "secret_ref_v1 must use kind env and a safe name")
	}
	if value, ok := lookupEnv(env, secretRefEnvName(name)); !ok || !isValidResolvedSecret(value) {
		return secretRef{}, providerConfigError(path, "provider_manifest_secret_missing", "secret_ref_v1 could not be resolved to a safe value")
	}
	return secretRef{Kind: kind, Name: name}, nil
}

func optionalAdditionalScopes(object map[string]json.RawMessage, index int) ([]string, error) {
	raw, ok := object["additional_scopes"]
	if !ok {
		return []string{}, nil
	}
	if isJSONNull(raw) {
		return nil, providerConfigError(providerPath(index, "additional_scopes"), "provider_manifest_schema_invalid", "additional_scopes must not be null")
	}
	var scopes []string
	if err := json.Unmarshal(raw, &scopes); err != nil || scopes == nil {
		return nil, providerConfigError(providerPath(index, "additional_scopes"), "provider_manifest_schema_invalid", "additional_scopes must be an array of strings")
	}
	if len(scopes) > 16 {
		return nil, providerConfigError(providerPath(index, "additional_scopes"), "provider_manifest_schema_invalid", "additional_scopes must contain at most 16 entries")
	}
	seen := make(map[string]struct{}, len(scopes))
	for scopeIndex, scope := range scopes {
		path := fmt.Sprintf("%s[%d]", providerPath(index, "additional_scopes"), scopeIndex)
		if scope == "" || scope == "openid" || !oidcScopeTokenPattern.MatchString(scope) {
			return nil, providerConfigError(path, "provider_manifest_schema_invalid", "additional scope is outside the accepted OIDC scope token contract")
		}
		if _, ok := seen[scope]; ok {
			return nil, providerConfigError(path, "provider_manifest_schema_invalid", "additional scope tokens must be unique")
		}
		seen[scope] = struct{}{}
	}
	return scopes, nil
}

func requiredSigningCertificates(object map[string]json.RawMessage, index int, reader DocumentReader) ([]string, error) {
	raw, ok := object["idp_signing_certificate_paths"]
	if !ok || isJSONNull(raw) {
		return nil, providerConfigError(providerPath(index, "idp_signing_certificate_paths"), "provider_manifest_schema_invalid", "idp_signing_certificate_paths is required")
	}
	var paths []string
	if err := json.Unmarshal(raw, &paths); err != nil || paths == nil {
		return nil, providerConfigError(providerPath(index, "idp_signing_certificate_paths"), "provider_manifest_schema_invalid", "idp_signing_certificate_paths must be an array of strings")
	}
	if len(paths) == 0 || len(paths) > 8 {
		return nil, providerConfigError(providerPath(index, "idp_signing_certificate_paths"), "provider_manifest_schema_invalid", "idp_signing_certificate_paths must contain 1..8 entries")
	}
	certificates := make([]string, 0, len(paths))
	for pathIndex, candidate := range paths {
		fieldPath := fmt.Sprintf("%s[%d]", providerPath(index, "idp_signing_certificate_paths"), pathIndex)
		if !validProviderManifestPath(candidate) {
			return nil, providerConfigError(fieldPath, "provider_manifest_referenced_file_invalid", "SAML signing certificate path must be an absolute POSIX path without forbidden segments")
		}
		rawCert, failure := reader.ReadDocument(candidate, SigningCertificateMaximumSize)
		if failure != "" {
			message := "SAML signing certificate is unavailable"
			if failure == DocumentTooLarge {
				message = "SAML signing certificate must not exceed 262144 bytes"
			} else if failure == DocumentUnsafe {
				message = "SAML signing certificate path must reference one regular file"
			}
			return nil, providerConfigError(fieldPath, "provider_manifest_referenced_file_invalid", message)
		}
		certificate, err := parseSigningCertificate(rawCert)
		if err != nil {
			return nil, providerConfigError(fieldPath, "provider_manifest_referenced_file_invalid", "SAML signing certificate is not parseable")
		}
		certificates = append(certificates, base64.StdEncoding.EncodeToString(certificate.Raw))
	}
	return certificates, nil
}

func requiredSAMLSubjectSource(object map[string]json.RawMessage, index int) (*authn.EnterpriseAuthSAMLSubjectSource, error) {
	raw, ok := object["subject_source"]
	if !ok || isJSONNull(raw) {
		return nil, providerConfigError(providerPath(index, "subject_source"), "provider_manifest_schema_invalid", "subject_source is required")
	}
	var sourceMap map[string]json.RawMessage
	if err := json.Unmarshal(raw, &sourceMap); err != nil || sourceMap == nil {
		return nil, providerConfigError(providerPath(index, "subject_source"), "provider_manifest_schema_invalid", "subject_source must be an object")
	}
	kind, err := requiredStringMember(sourceMap, providerPath(index, "subject_source.kind"))
	if err != nil {
		return nil, providerConfigError(providerPath(index, "subject_source.kind"), "provider_manifest_schema_invalid", "subject_source.kind is required")
	}
	switch kind {
	case "name_id":
		for key := range sourceMap {
			if key != "kind" {
				return nil, providerConfigError(providerPath(index, "subject_source."+key), "provider_manifest_schema_invalid", "unknown subject_source member")
			}
		}
		return &authn.EnterpriseAuthSAMLSubjectSource{Kind: "name_id"}, nil
	case "attribute":
		for key := range sourceMap {
			if key != "kind" && key != "attribute_name" {
				return nil, providerConfigError(providerPath(index, "subject_source."+key), "provider_manifest_schema_invalid", "unknown subject_source member")
			}
		}
		attributeName, err := requiredStringMember(sourceMap, providerPath(index, "subject_source.attribute_name"))
		if err != nil || strings.TrimSpace(attributeName) == "" {
			return nil, providerConfigError(providerPath(index, "subject_source.attribute_name"), "provider_manifest_schema_invalid", "subject_source.attribute_name must be non-empty")
		}
		return &authn.EnterpriseAuthSAMLSubjectSource{Kind: "attribute", AttributeName: attributeName}, nil
	default:
		return nil, providerConfigError(providerPath(index, "subject_source.kind"), "provider_manifest_schema_invalid", "subject_source.kind must be name_id or attribute")
	}
}

func rejectDuplicateObjectMembers(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := scanJSONValueForDuplicateMembers(decoder); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("provider manifest must not be empty")
		}
		return err
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("provider manifest contains trailing JSON token %v", token)
	}
	return nil
}

func scanJSONValueForDuplicateMembers(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok {
				return fmt.Errorf("provider manifest object member is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("%w %q", errDuplicateJSONMember, key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValueForDuplicateMembers(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := scanJSONValueForDuplicateMembers(decoder); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

func parseSigningCertificate(raw []byte) (*x509.Certificate, error) {
	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			return x509.ParseCertificate(block.Bytes)
		}
		raw = rest
	}
	return x509.ParseCertificate(bytes.TrimSpace(raw))
}

func validHTTPSProviderURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return false
	}
	return parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == ""
}

func validProviderManifestPath(raw string) bool {
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.ContainsRune(raw, '\x00') || strings.HasPrefix(raw, "~") || strings.Contains(raw, "$") {
		return false
	}
	for _, segment := range strings.Split(raw, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return pathpkg.Clean(raw) == raw
}

func isJSONNull(raw json.RawMessage) bool {
	return strings.TrimSpace(string(raw)) == "null"
}

func providerBasePath(index int) string {
	return fmt.Sprintf("%s.providers[%d]", providerManifestRootPath, index)
}

func providerPath(index int, field string) string {
	return providerBasePath(index) + "." + field
}

func stringPtr(value string) *string {
	return &value
}

func isValidSecretRefName(value string) bool {
	if !secretRefNamePattern.MatchString(value) {
		return false
	}
	return normalizedSecretRefSuffix(value) != ""
}

func isValidResolvedSecret(value string) bool {
	if value == "" || strings.TrimSpace(value) != value || !utf8.ValidString(value) || len(value) > 4096 {
		return false
	}
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\x00' || r == 0x7f || r < 0x20 || r > 0x7e {
			return false
		}
	}
	return true
}

func hasControlRunes(value string) bool {
	for _, r := range value {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return true
		}
	}
	return false
}
