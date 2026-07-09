package authn

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const EnterpriseAuthTransactionTTL = 10 * time.Minute

func (s *Store) ListEnterpriseAuthProviders(ctx context.Context) ([]EnterpriseAuthProviderRecord, error) {
	rows, err := s.pool.Query(ctx, `
SELECT id, provider_key, provider_type, display_name, is_enabled, is_interactive,
       authorization_endpoint, issuer, audience, token_endpoint, jwks_uri, client_id,
       client_secret_ref_kind, client_secret_ref_name, additional_scopes,
       saml_idp_entity_id, saml_sso_url, saml_idp_signing_certificates,
       saml_sp_entity_id, saml_subject_source, created_at, updated_at
  FROM enterprise_auth_providers
 WHERE is_enabled = true
   AND is_interactive = true
 ORDER BY display_name ASC, provider_key ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list enterprise auth providers: %w", err)
	}
	defer rows.Close()

	providers := make([]EnterpriseAuthProviderRecord, 0)
	for rows.Next() {
		provider, err := scanEnterpriseAuthProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enterprise auth providers: %w", err)
	}
	return providers, nil
}

func (s *Store) GetEnterpriseAuthProviderByKey(ctx context.Context, providerKey string) (EnterpriseAuthProviderRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, provider_key, provider_type, display_name, is_enabled, is_interactive,
       authorization_endpoint, issuer, audience, token_endpoint, jwks_uri, client_id,
       client_secret_ref_kind, client_secret_ref_name, additional_scopes,
       saml_idp_entity_id, saml_sso_url, saml_idp_signing_certificates,
       saml_sp_entity_id, saml_subject_source, created_at, updated_at
  FROM enterprise_auth_providers
 WHERE provider_key = $1
`, providerKey)
	provider, err := scanEnterpriseAuthProvider(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthProviderRecord{}, ErrAuthProviderNotFound
	}
	return provider, err
}

func (s *Store) CreateEnterpriseAuthTransaction(
	ctx context.Context,
	provider EnterpriseAuthProviderRecord,
	returnTo string,
	state *string,
	nonce *string,
	pkceVerifierHash []byte,
	pkceVerifierCiphertext []byte,
	pkceVerifierNonce []byte,
	relayState *string,
	samlRequestID *string,
	browserBindingHash []byte,
	now time.Time,
) (EnterpriseAuthTransactionRecord, error) {
	var record EnterpriseAuthTransactionRecord
	if err := s.pool.QueryRow(ctx, `
INSERT INTO enterprise_auth_transactions (
    provider_id, provider_key, provider_type, return_to, state, nonce, pkce_verifier_hash,
    pkce_verifier_ciphertext, pkce_verifier_nonce, relay_state, saml_request_id,
    browser_binding_hash, created_at, expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
          browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
          saml_completion_hash, saml_subject, saml_staged_at,
          created_at, expires_at, consumed_at
`, provider.ID, provider.ProviderKey, provider.ProviderType, returnTo, state, nonce, pkceVerifierHash, pkceVerifierCiphertext, pkceVerifierNonce, relayState, samlRequestID, browserBindingHash, now.UTC(), now.UTC().Add(EnterpriseAuthTransactionTTL)).Scan(
		&record.ID,
		&record.ProviderID,
		&record.ProviderKey,
		&record.ProviderType,
		&record.ReturnTo,
		&record.State,
		&record.Nonce,
		&record.RelayState,
		&record.BrowserBindingHash,
		&record.PKCEVerifierCiphertext,
		&record.PKCEVerifierNonce,
		&record.SAMLRequestID,
		&record.SAMLCompletionHash,
		&record.SAMLSubject,
		&record.SAMLStagedAt,
		&record.CreatedAt,
		&record.ExpiresAt,
		&record.ConsumedAt,
	); err != nil {
		return EnterpriseAuthTransactionRecord{}, fmt.Errorf("insert enterprise auth transaction: %w", err)
	}
	return record, nil
}

func (s *Store) ReconcileEnterpriseAuthProviders(ctx context.Context, definitions []EnterpriseAuthProviderDefinition, now time.Time) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin enterprise auth provider reconciliation: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	existingRows, err := tx.Query(ctx, `
SELECT provider_key, provider_type
  FROM enterprise_auth_providers
 FOR UPDATE
`)
	if err != nil {
		return fmt.Errorf("lock enterprise auth providers: %w", err)
	}
	existing := make(map[string]string)
	for existingRows.Next() {
		var providerKey, providerType string
		if err := existingRows.Scan(&providerKey, &providerType); err != nil {
			existingRows.Close()
			return fmt.Errorf("scan enterprise auth provider lock: %w", err)
		}
		existing[providerKey] = providerType
	}
	if err := existingRows.Err(); err != nil {
		existingRows.Close()
		return fmt.Errorf("iterate enterprise auth provider locks: %w", err)
	}
	existingRows.Close()

	ordered := append([]EnterpriseAuthProviderDefinition(nil), definitions...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ProviderKey < ordered[j].ProviderKey
	})

	manifestKeys := make(map[string]struct{}, len(ordered))
	for _, definition := range ordered {
		manifestKeys[definition.ProviderKey] = struct{}{}
		if existingType, ok := existing[definition.ProviderKey]; ok && existingType != definition.ProviderType {
			return ErrAuthProviderTypeChangeNotSupported
		}

		additionalScopesJSON, err := encodeStringArrayJSON(definition.AdditionalScopes)
		if err != nil {
			return fmt.Errorf("encode enterprise auth provider additional scopes: %w", err)
		}
		samlSigningCertificatesJSON, err := encodeStringArrayJSON(definition.SAMLIDPSigningCertificate)
		if err != nil {
			return fmt.Errorf("encode enterprise auth provider SAML signing certificates: %w", err)
		}
		samlSubjectSourceJSON, err := encodeOptionalSAMLSubjectSourceJSON(definition.SAMLSubjectSource)
		if err != nil {
			return fmt.Errorf("encode enterprise auth provider SAML subject source: %w", err)
		}

		if _, err := tx.Exec(ctx, `
INSERT INTO enterprise_auth_providers (
    provider_key, provider_type, display_name, is_enabled, is_interactive,
    authorization_endpoint, issuer, audience, token_endpoint, jwks_uri, client_id,
    client_secret_ref_kind, client_secret_ref_name, additional_scopes,
    saml_idp_entity_id, saml_sso_url, saml_idp_signing_certificates,
    saml_sp_entity_id, saml_subject_source, updated_at
)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14, $15, $16::jsonb, $17, $18::jsonb, $19)
ON CONFLICT (provider_key) DO UPDATE
   SET display_name = EXCLUDED.display_name,
       is_enabled = EXCLUDED.is_enabled,
       is_interactive = EXCLUDED.is_interactive,
       authorization_endpoint = EXCLUDED.authorization_endpoint,
       issuer = EXCLUDED.issuer,
       audience = EXCLUDED.audience,
       token_endpoint = EXCLUDED.token_endpoint,
       jwks_uri = EXCLUDED.jwks_uri,
       client_id = EXCLUDED.client_id,
       client_secret_ref_kind = EXCLUDED.client_secret_ref_kind,
       client_secret_ref_name = EXCLUDED.client_secret_ref_name,
       additional_scopes = EXCLUDED.additional_scopes,
       saml_idp_entity_id = EXCLUDED.saml_idp_entity_id,
       saml_sso_url = EXCLUDED.saml_sso_url,
       saml_idp_signing_certificates = EXCLUDED.saml_idp_signing_certificates,
       saml_sp_entity_id = EXCLUDED.saml_sp_entity_id,
       saml_subject_source = EXCLUDED.saml_subject_source,
       updated_at = EXCLUDED.updated_at
`, definition.ProviderKey, definition.ProviderType, definition.DisplayName, definition.Enabled,
			definition.AuthorizationEndpoint, definition.Issuer, definition.Audience, definition.TokenEndpoint, definition.JWKSURI, definition.ClientID,
			definition.ClientSecretRefKind, definition.ClientSecretRefName, additionalScopesJSON,
			definition.SAMLIDPEntityID, definition.SAMLSSOURL, samlSigningCertificatesJSON,
			definition.SAMLSPHostEntityID, samlSubjectSourceJSON, now.UTC()); err != nil {
			return fmt.Errorf("upsert enterprise auth provider %q: %w", definition.ProviderKey, err)
		}
	}

	if len(manifestKeys) < len(existing) {
		omitted := make([]string, 0)
		for providerKey := range existing {
			if _, ok := manifestKeys[providerKey]; !ok {
				omitted = append(omitted, providerKey)
			}
		}
		sort.Strings(omitted)
		for _, providerKey := range omitted {
			if _, err := tx.Exec(ctx, `
UPDATE enterprise_auth_providers
   SET is_enabled = false,
       is_interactive = false,
       updated_at = $2
 WHERE provider_key = $1
`, providerKey, now.UTC()); err != nil {
				return fmt.Errorf("disable omitted enterprise auth provider %q: %w", providerKey, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit enterprise auth provider reconciliation: %w", err)
	}
	return nil
}

func (s *Store) GetOIDCEnterpriseAuthTransactionForCallback(ctx context.Context, providerKey string, state string, browserBindingHash []byte, now time.Time) (EnterpriseAuthTransactionRecord, error) {
	record, err := scanEnterpriseAuthTransaction(s.pool.QueryRow(ctx, `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE provider_key = $1
   AND provider_type = 'oidc'
   AND state = $2
`, providerKey, state))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionNotFound
	}
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, err
	}
	if record.ConsumedAt != nil {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionUsed
	}
	if !record.ExpiresAt.After(now.UTC()) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionExpired
	}
	if !bytes.Equal(record.BrowserBindingHash, browserBindingHash) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionBrowserMismatch
	}
	return record, nil
}

func (s *Store) GetSAMLEnterpriseAuthTransactionForACS(ctx context.Context, providerKey string, relayState string, now time.Time) (EnterpriseAuthTransactionRecord, error) {
	record, err := scanEnterpriseAuthTransaction(s.pool.QueryRow(ctx, `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE provider_key = $1
   AND provider_type = 'saml'
   AND relay_state = $2
`, providerKey, relayState))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionNotFound
	}
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, err
	}
	if record.ConsumedAt != nil || record.SAMLCompletionHash != nil || record.SAMLSubject != nil || record.SAMLStagedAt != nil {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionUsed
	}
	if !record.ExpiresAt.After(now.UTC()) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionExpired
	}
	return record, nil
}

func (s *Store) CompleteOIDCEnterpriseAuthTransaction(
	ctx context.Context,
	providerKey string,
	state string,
	browserBindingHash []byte,
	nonce *string,
	providerSubject string,
	now time.Time,
) (EnterpriseAuthCompletionResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthCompletionResult{}, fmt.Errorf("begin enterprise auth completion: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	provider, err := fetchEnterpriseAuthProviderByKeyTx(ctx, tx, providerKey)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if provider.ProviderType != "oidc" {
		return EnterpriseAuthCompletionResult{}, ErrAuthProviderNotFound
	}
	if !provider.IsEnabled {
		return EnterpriseAuthCompletionResult{}, ErrAuthProviderDisabled
	}

	transaction, err := fetchEnterpriseTransactionByBrowserBindingTx(ctx, tx, "oidc", browserBindingHash)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if err := validateEnterpriseTransactionForProvider(provider, transaction, "oidc", now); err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if transaction.State == nil || *transaction.State != state {
		return EnterpriseAuthCompletionResult{}, ErrEnterpriseTransactionStateMismatch
	}
	if transaction.Nonce == nil || nonce == nil || *transaction.Nonce != *nonce {
		return EnterpriseAuthCompletionResult{}, ErrSubjectMismatch
	}

	return completeEnterpriseAuthTransaction(ctx, tx, transaction, provider.ID, providerSubject, now)
}

func (s *Store) StageSAMLEnterpriseAuthTransaction(
	ctx context.Context,
	providerKey string,
	relayState string,
	providerSubject string,
	completionHash []byte,
	now time.Time,
) (EnterpriseAuthTransactionRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, fmt.Errorf("begin enterprise auth SAML staging: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	provider, err := fetchEnterpriseAuthProviderByKeyTx(ctx, tx, providerKey)
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, err
	}
	if provider.ProviderType != "saml" {
		return EnterpriseAuthTransactionRecord{}, ErrAuthProviderNotFound
	}
	if !provider.IsEnabled {
		return EnterpriseAuthTransactionRecord{}, ErrAuthProviderDisabled
	}

	transaction, err := fetchEnterpriseTransactionByCorrelationTx(ctx, tx, "saml", relayState)
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, err
	}
	if err := validateEnterpriseTransactionForProvider(provider, transaction, "saml", now); err != nil {
		return EnterpriseAuthTransactionRecord{}, err
	}
	if transaction.SAMLCompletionHash != nil || transaction.SAMLSubject != nil || transaction.SAMLStagedAt != nil {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionUsed
	}

	stagedAt := now.UTC()
	staged, err := scanEnterpriseAuthTransaction(tx.QueryRow(ctx, `
UPDATE enterprise_auth_transactions
   SET saml_completion_hash = $2,
       saml_subject = $3,
       saml_staged_at = $4
 WHERE id = $1
RETURNING id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
          browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
          saml_completion_hash, saml_subject, saml_staged_at,
          created_at, expires_at, consumed_at
`, transaction.ID, completionHash, providerSubject, stagedAt))
	if err != nil {
		return EnterpriseAuthTransactionRecord{}, fmt.Errorf("stage enterprise auth SAML transaction: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return EnterpriseAuthTransactionRecord{}, fmt.Errorf("commit enterprise auth SAML staging: %w", err)
	}
	return staged, nil
}

func (s *Store) CompleteSAMLEnterpriseAuthTransaction(
	ctx context.Context,
	providerKey string,
	completionHash []byte,
	browserBindingHash []byte,
	now time.Time,
) (EnterpriseAuthCompletionResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthCompletionResult{}, fmt.Errorf("begin enterprise auth SAML completion: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	provider, err := fetchEnterpriseAuthProviderByKeyTx(ctx, tx, providerKey)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if provider.ProviderType != "saml" {
		return EnterpriseAuthCompletionResult{}, ErrAuthProviderNotFound
	}
	if !provider.IsEnabled {
		return EnterpriseAuthCompletionResult{}, ErrAuthProviderDisabled
	}

	transaction, err := fetchEnterpriseTransactionByCompletionHashTx(ctx, tx, completionHash)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if err := validateEnterpriseTransactionForProvider(provider, transaction, "saml", now); err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if !bytes.Equal(transaction.BrowserBindingHash, browserBindingHash) {
		return EnterpriseAuthCompletionResult{}, ErrEnterpriseTransactionBrowserMismatch
	}
	if transaction.SAMLSubject == nil || *transaction.SAMLSubject == "" {
		return EnterpriseAuthCompletionResult{}, ErrEnterpriseTransactionCompletionMismatch
	}

	return completeEnterpriseAuthTransaction(ctx, tx, transaction, provider.ID, *transaction.SAMLSubject, now)
}

func validateEnterpriseTransactionForProvider(provider EnterpriseAuthProviderRecord, transaction EnterpriseAuthTransactionRecord, providerType string, now time.Time) error {
	if transaction.ProviderID != provider.ID || transaction.ProviderKey != provider.ProviderKey || transaction.ProviderType != providerType {
		return ErrEnterpriseTransactionProviderMismatch
	}
	if transaction.ConsumedAt != nil {
		return ErrEnterpriseTransactionUsed
	}
	if !transaction.ExpiresAt.After(now.UTC()) {
		return ErrEnterpriseTransactionExpired
	}
	return nil
}

func completeEnterpriseAuthTransaction(ctx context.Context, tx pgx.Tx, transaction EnterpriseAuthTransactionRecord, providerID uuid.UUID, providerSubject string, now time.Time) (EnterpriseAuthCompletionResult, error) {
	binding, err := fetchActiveEnterpriseBindingBySubjectTx(ctx, tx, providerID, providerSubject)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	user, err := fetchUserForUpdate(ctx, tx, binding.UserID)
	if err != nil {
		return EnterpriseAuthCompletionResult{}, err
	}
	if !user.IsActive {
		return EnterpriseAuthCompletionResult{}, ErrEnterpriseIdentityInactiveUser
	}

	completedAt := now.UTC()
	if _, err := tx.Exec(ctx, `
UPDATE enterprise_auth_transactions
   SET consumed_at = $2
 WHERE id = $1
`, transaction.ID, completedAt); err != nil {
		return EnterpriseAuthCompletionResult{}, fmt.Errorf("consume enterprise auth transaction: %w", err)
	}
	if err := tx.QueryRow(ctx, `
UPDATE enterprise_auth_bindings
   SET last_auth_at = $2
 WHERE id = $1
RETURNING id, user_id, provider_key, provider_type, provider_subject, created_at, last_auth_at
`, binding.ID, completedAt).Scan(
		&binding.ID,
		&binding.UserID,
		&binding.ProviderKey,
		&binding.ProviderType,
		&binding.ProviderSubject,
		&binding.CreatedAt,
		&binding.LastAuthAt,
	); err != nil {
		return EnterpriseAuthCompletionResult{}, fmt.Errorf("update enterprise binding last_auth_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return EnterpriseAuthCompletionResult{}, fmt.Errorf("commit enterprise auth completion: %w", err)
	}
	return EnterpriseAuthCompletionResult{
		User:          user,
		Binding:       binding,
		ReturnTo:      transaction.ReturnTo,
		TransactionID: transaction.ID,
	}, nil
}

func (s *Store) ListEnterpriseAuthBindingSummaries(ctx context.Context, userID uuid.UUID) ([]EnterpriseAuthBindingSummary, error) {
	return listEnterpriseAuthBindingSummaries(ctx, s.pool, userID)
}

func (s *Store) CreateEnterpriseAuthBinding(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	baseUserVersion int64,
	clientTxnID string,
	providerKey string,
	providerSubject string,
	reason *string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (EnterpriseAuthBindingResult, error) {
	key := RouteIdempotencyKey{
		RouteKey:    "users.auth_bindings.create",
		ActorUserID: actor.ID,
		ScopeKey:    targetUserID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{ResponseJSON: existing.ResponseJSON, Replayed: true, StatusCode: http.StatusOK}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return EnterpriseAuthBindingResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if target.UserVersion != baseUserVersion {
		return EnterpriseAuthBindingResult{}, ErrUserVersionConflict
	}
	provider, err := fetchEnterpriseAuthProviderByKeyTx(ctx, tx, providerKey)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if hasActiveEnterpriseBindingForUserProviderTx(ctx, tx, targetUserID, provider.ID) {
		return EnterpriseAuthBindingResult{}, ErrAuthBindingProviderAlreadyLinkedForUser
	}
	if hasActiveEnterpriseBindingForProviderSubjectTx(ctx, tx, provider.ID, providerSubject) {
		return EnterpriseAuthBindingResult{}, ErrAuthBindingProviderSubjectInUse
	}

	changedAt := now.UTC()
	var bindingID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO enterprise_auth_bindings (
    user_id, provider_id, provider_key, provider_type, provider_subject, created_by_user_id, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id
`, targetUserID, provider.ID, provider.ProviderKey, provider.ProviderType, providerSubject, actor.ID, changedAt).Scan(&bindingID); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}

	updated, err := updateUserVersionTx(ctx, tx, actor.ID, targetUserID, changedAt)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	bindings, err := listEnterpriseAuthBindingSummaries(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	responseJSON, err := json.Marshal(SafeUserResponseWithEnterpriseBindings(updated, bindings))
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}

	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusCreated, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, after_json)
VALUES ($1, $2, 'users.auth_bindings.create', 'auth_binding_created', 'auth_binding_created', $3, $4,
        jsonb_build_object('auth_binding_id', $5::text, 'provider_key', $6::text, 'provider_subject', $7::text, 'reason', $8::text))
`, actor.ID, targetUserID, clientTxnID, requestID, bindingID.String(), provider.ProviderKey, providerSubject, nullableText(reason)); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	return EnterpriseAuthBindingResult{User: updated, ResponseJSON: responseJSON, StatusCode: http.StatusCreated}, nil
}

func (s *Store) RotateEnterpriseAuthBinding(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	authBindingID uuid.UUID,
	baseUserVersion int64,
	clientTxnID string,
	newProviderSubject string,
	reason *string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (EnterpriseAuthBindingResult, error) {
	key := RouteIdempotencyKey{
		RouteKey:    "users.auth_bindings.rotate",
		ActorUserID: actor.ID,
		ScopeKey:    authBindingID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{ResponseJSON: existing.ResponseJSON, Replayed: true, StatusCode: http.StatusOK}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return EnterpriseAuthBindingResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if target.UserVersion != baseUserVersion {
		return EnterpriseAuthBindingResult{}, ErrUserVersionConflict
	}
	current, err := fetchEnterpriseBindingForUpdateTx(ctx, tx, targetUserID, authBindingID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if current.RetiredAt.Valid {
		return EnterpriseAuthBindingResult{}, ErrAuthBindingNotActive
	}

	changedAt := now.UTC()
	if current.ProviderSubject == newProviderSubject {
		bindings, err := listEnterpriseAuthBindingSummaries(ctx, tx, targetUserID)
		if err != nil {
			return EnterpriseAuthBindingResult{}, err
		}
		responseJSON, err := json.Marshal(SafeUserResponseWithEnterpriseBindings(target, bindings))
		if err != nil {
			return EnterpriseAuthBindingResult{}, err
		}
		if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
			if IsUniqueViolation(err) {
				return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
			}
			return EnterpriseAuthBindingResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return EnterpriseAuthBindingResult{}, err
		}
		return EnterpriseAuthBindingResult{User: target, ResponseJSON: responseJSON, StatusCode: http.StatusOK}, nil
	}
	if hasActiveEnterpriseBindingForProviderSubjectTx(ctx, tx, current.ProviderID, newProviderSubject) {
		return EnterpriseAuthBindingResult{}, ErrAuthBindingProviderSubjectInUse
	}

	if _, err := tx.Exec(ctx, `
	UPDATE enterprise_auth_bindings
	   SET retired_at = $2,
	       retired_by_user_id = $3,
	       retire_reason = $4
	 WHERE id = $1
	`, authBindingID, changedAt, actor.ID, nullableText(reason)); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	var replacementID uuid.UUID
	if err := tx.QueryRow(ctx, `
	INSERT INTO enterprise_auth_bindings (
	    user_id, provider_id, provider_key, provider_type, provider_subject, created_by_user_id, created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`, targetUserID, current.ProviderID, current.ProviderKey, current.ProviderType, newProviderSubject, actor.ID, changedAt).Scan(&replacementID); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if _, err := tx.Exec(ctx, `
	UPDATE enterprise_auth_bindings
	   SET replaced_by_auth_binding_id = $2
	 WHERE id = $1
	`, authBindingID, replacementID); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	updated, err := updateUserVersionTx(ctx, tx, actor.ID, targetUserID, changedAt)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	revoked, err := revokeAllSessionsTx(ctx, tx, targetUserID, changedAt)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	bindings, err := listEnterpriseAuthBindingSummaries(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	responseJSON, err := json.Marshal(SafeUserResponseWithEnterpriseBindings(updated, bindings))
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, before_json, after_json)
VALUES ($1, $2, 'users.auth_bindings.rotate', 'auth_binding_rotated', 'auth_binding_rotated', $3, $4,
        jsonb_build_object('auth_binding_id', $5::text, 'provider_key', $6::text, 'provider_subject', $7::text),
        jsonb_build_object('auth_binding_id', $8::text, 'provider_key', $6::text, 'provider_subject', $9::text, 'reason', $10::text))
`, actor.ID, targetUserID, clientTxnID, requestID, authBindingID.String(), current.ProviderKey, current.ProviderSubject, replacementID.String(), newProviderSubject, nullableText(reason)); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	return EnterpriseAuthBindingResult{User: updated, ResponseJSON: responseJSON, RevokedSessionIDs: revoked, StatusCode: http.StatusOK}, nil
}

func (s *Store) RetireEnterpriseAuthBinding(
	ctx context.Context,
	actor UserRecord,
	targetUserID uuid.UUID,
	authBindingID uuid.UUID,
	baseUserVersion int64,
	clientTxnID string,
	reason *string,
	requestHash []byte,
	requestID string,
	now time.Time,
) (EnterpriseAuthBindingResult, error) {
	key := RouteIdempotencyKey{
		RouteKey:    "users.auth_bindings.retire",
		ActorUserID: actor.ID,
		ScopeKey:    authBindingID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.GetRouteIdempotency(ctx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{ResponseJSON: existing.ResponseJSON, Replayed: true, StatusCode: http.StatusOK}, nil
	} else if !errors.Is(err, ErrNotFound) {
		return EnterpriseAuthBindingResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	target, err := fetchUserForUpdate(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if target.UserVersion != baseUserVersion {
		return EnterpriseAuthBindingResult{}, ErrUserVersionConflict
	}
	current, err := fetchEnterpriseBindingForUpdateTx(ctx, tx, targetUserID, authBindingID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if current.RetiredAt.Valid {
		return EnterpriseAuthBindingResult{}, ErrAuthBindingNotActive
	}

	changedAt := now.UTC()
	if _, err := tx.Exec(ctx, `
UPDATE enterprise_auth_bindings
   SET retired_at = $2,
       retired_by_user_id = $3,
       retire_reason = $4
 WHERE id = $1
`, authBindingID, changedAt, actor.ID, nullableText(reason)); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	updated, err := updateUserVersionTx(ctx, tx, actor.ID, targetUserID, changedAt)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	revoked, err := revokeAllSessionsTx(ctx, tx, targetUserID, changedAt)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	bindings, err := listEnterpriseAuthBindingSummaries(ctx, tx, targetUserID)
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	responseJSON, err := json.Marshal(SafeUserResponseWithEnterpriseBindings(updated, bindings))
	if err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if err := InsertRouteIdempotency(ctx, tx, key, &targetUserID, requestHash, http.StatusOK, responseJSON); err != nil {
		if IsUniqueViolation(err) {
			return EnterpriseAuthBindingResult{}, ErrClientTxnConflict
		}
		return EnterpriseAuthBindingResult{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (actor_user_id, target_user_id, event_source, event_kind, reason_code, client_txn_id, request_id, before_json, after_json)
VALUES ($1, $2, 'users.auth_bindings.retire', 'auth_binding_retired', 'auth_binding_retired', $3, $4,
        jsonb_build_object('auth_binding_id', $5::text, 'provider_key', $6::text, 'provider_subject', $7::text),
        jsonb_build_object('retired_at', $8::timestamptz, 'reason', $9::text))
`, actor.ID, targetUserID, clientTxnID, requestID, authBindingID.String(), current.ProviderKey, current.ProviderSubject, changedAt, nullableText(reason)); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return EnterpriseAuthBindingResult{}, err
	}
	return EnterpriseAuthBindingResult{User: updated, ResponseJSON: responseJSON, RevokedSessionIDs: revoked, StatusCode: http.StatusOK}, nil
}

func SafeUserResponseWithEnterpriseBindings(user UserRecord, bindings []EnterpriseAuthBindingSummary) map[string]any {
	authBindings := make([]map[string]any, 0, len(bindings)+1)
	authBindings = append(authBindings, map[string]any{
		"provider_type": "local",
		"provider_key":  "local",
		"username":      user.Email,
		"created_at":    user.CreatedAt,
	})
	for _, binding := range bindings {
		authBindings = append(authBindings, map[string]any{
			"auth_binding_id":  binding.ID,
			"provider_key":     binding.ProviderKey,
			"provider_type":    binding.ProviderType,
			"provider_subject": binding.ProviderSubject,
			"created_at":       binding.CreatedAt,
			"last_auth_at":     binding.LastAuthAt,
		})
	}

	payload := safeUserResponse(user)
	payload["auth_bindings"] = authBindings
	return payload
}

type enterpriseBindingRow struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProviderID      uuid.UUID
	ProviderKey     string
	ProviderType    string
	ProviderSubject string
	CreatedAt       time.Time
	LastAuthAt      *time.Time
	RetiredAt       sql.NullTime
}

type enterpriseBindingScanner interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func scanEnterpriseAuthProvider(scanner interface{ Scan(...any) error }) (EnterpriseAuthProviderRecord, error) {
	var provider EnterpriseAuthProviderRecord
	var additionalScopesRaw []byte
	var samlSigningCertificatesRaw []byte
	var samlSubjectSourceRaw []byte
	err := scanner.Scan(
		&provider.ID,
		&provider.ProviderKey,
		&provider.ProviderType,
		&provider.DisplayName,
		&provider.IsEnabled,
		&provider.IsInteractive,
		&provider.AuthorizationEndpoint,
		&provider.Issuer,
		&provider.Audience,
		&provider.TokenEndpoint,
		&provider.JWKSURI,
		&provider.ClientID,
		&provider.ClientSecretRefKind,
		&provider.ClientSecretRefName,
		&additionalScopesRaw,
		&provider.SAMLIDPEntityID,
		&provider.SAMLSSOURL,
		&samlSigningCertificatesRaw,
		&provider.SAMLSPHostEntityID,
		&samlSubjectSourceRaw,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		return EnterpriseAuthProviderRecord{}, err
	}
	provider.AdditionalScopes, err = decodeStringArrayJSON(additionalScopesRaw)
	if err != nil {
		return EnterpriseAuthProviderRecord{}, fmt.Errorf("decode enterprise auth provider additional scopes: %w", err)
	}
	provider.SAMLIDPSigningCertificate, err = decodeStringArrayJSON(samlSigningCertificatesRaw)
	if err != nil {
		return EnterpriseAuthProviderRecord{}, fmt.Errorf("decode enterprise auth provider SAML signing certificates: %w", err)
	}
	provider.SAMLSubjectSource, err = decodeOptionalSAMLSubjectSourceJSON(samlSubjectSourceRaw)
	if err != nil {
		return EnterpriseAuthProviderRecord{}, fmt.Errorf("decode enterprise auth provider SAML subject source: %w", err)
	}
	return provider, nil
}

func scanEnterpriseAuthTransaction(scanner interface{ Scan(...any) error }) (EnterpriseAuthTransactionRecord, error) {
	var record EnterpriseAuthTransactionRecord
	err := scanner.Scan(
		&record.ID,
		&record.ProviderID,
		&record.ProviderKey,
		&record.ProviderType,
		&record.ReturnTo,
		&record.State,
		&record.Nonce,
		&record.RelayState,
		&record.BrowserBindingHash,
		&record.PKCEVerifierCiphertext,
		&record.PKCEVerifierNonce,
		&record.SAMLRequestID,
		&record.SAMLCompletionHash,
		&record.SAMLSubject,
		&record.SAMLStagedAt,
		&record.CreatedAt,
		&record.ExpiresAt,
		&record.ConsumedAt,
	)
	return record, err
}

func fetchEnterpriseAuthProviderByKeyTx(ctx context.Context, tx pgx.Tx, providerKey string) (EnterpriseAuthProviderRecord, error) {
	provider, err := scanEnterpriseAuthProvider(tx.QueryRow(ctx, `
SELECT id, provider_key, provider_type, display_name, is_enabled, is_interactive,
       authorization_endpoint, issuer, audience, token_endpoint, jwks_uri, client_id,
       client_secret_ref_kind, client_secret_ref_name, additional_scopes,
       saml_idp_entity_id, saml_sso_url, saml_idp_signing_certificates,
       saml_sp_entity_id, saml_subject_source, created_at, updated_at
  FROM enterprise_auth_providers
 WHERE provider_key = $1
`, providerKey))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthProviderRecord{}, ErrAuthProviderNotFound
	}
	return provider, err
}

func fetchEnterpriseTransactionByCorrelationTx(ctx context.Context, tx pgx.Tx, providerType string, correlation string) (EnterpriseAuthTransactionRecord, error) {
	query := `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE state = $1
 FOR UPDATE`
	if providerType == "saml" {
		query = `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE relay_state = $1
 FOR UPDATE`
	}
	record, err := scanEnterpriseAuthTransaction(tx.QueryRow(ctx, query, correlation))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionNotFound
	}
	return record, err
}

func fetchEnterpriseTransactionByBrowserBindingTx(ctx context.Context, tx pgx.Tx, providerType string, browserBindingHash []byte) (EnterpriseAuthTransactionRecord, error) {
	record, err := scanEnterpriseAuthTransaction(tx.QueryRow(ctx, `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE provider_type = $1
   AND browser_binding_hash = $2
 FOR UPDATE
`, providerType, browserBindingHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionBrowserMismatch
	}
	return record, err
}

func fetchEnterpriseTransactionByCompletionHashTx(ctx context.Context, tx pgx.Tx, completionHash []byte) (EnterpriseAuthTransactionRecord, error) {
	record, err := scanEnterpriseAuthTransaction(tx.QueryRow(ctx, `
SELECT id, provider_id, provider_key, provider_type, return_to, state, nonce, relay_state,
       browser_binding_hash, pkce_verifier_ciphertext, pkce_verifier_nonce, saml_request_id,
       saml_completion_hash, saml_subject, saml_staged_at,
       created_at, expires_at, consumed_at
  FROM enterprise_auth_transactions
 WHERE saml_completion_hash = $1
 FOR UPDATE
`, completionHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return EnterpriseAuthTransactionRecord{}, ErrEnterpriseTransactionCompletionMismatch
	}
	return record, err
}

func fetchActiveEnterpriseBindingBySubjectTx(ctx context.Context, tx pgx.Tx, providerID uuid.UUID, providerSubject string) (EnterpriseAuthBindingSummary, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, provider_key, provider_type, provider_subject, created_at, last_auth_at
  FROM enterprise_auth_bindings
 WHERE provider_id = $1
   AND provider_subject = $2
   AND retired_at IS NULL
`, providerID, providerSubject)
	var binding EnterpriseAuthBindingSummary
	if err := row.Scan(&binding.ID, &binding.UserID, &binding.ProviderKey, &binding.ProviderType, &binding.ProviderSubject, &binding.CreatedAt, &binding.LastAuthAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EnterpriseAuthBindingSummary{}, ErrEnterpriseIdentityNoLinkedUser
		}
		return EnterpriseAuthBindingSummary{}, err
	}
	return binding, nil
}

func listEnterpriseAuthBindingSummaries(ctx context.Context, querier enterpriseBindingScanner, userID uuid.UUID) ([]EnterpriseAuthBindingSummary, error) {
	rows, err := querier.Query(ctx, `
SELECT id, user_id, provider_key, provider_type, provider_subject, created_at, last_auth_at
  FROM enterprise_auth_bindings
 WHERE user_id = $1
   AND retired_at IS NULL
 ORDER BY provider_type ASC, provider_key ASC, created_at ASC
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bindings := make([]EnterpriseAuthBindingSummary, 0)
	for rows.Next() {
		var binding EnterpriseAuthBindingSummary
		if err := rows.Scan(&binding.ID, &binding.UserID, &binding.ProviderKey, &binding.ProviderType, &binding.ProviderSubject, &binding.CreatedAt, &binding.LastAuthAt); err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}

func fetchEnterpriseBindingForUpdateTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, authBindingID uuid.UUID) (enterpriseBindingRow, error) {
	row := tx.QueryRow(ctx, `
SELECT id, user_id, provider_id, provider_key, provider_type, provider_subject, created_at, last_auth_at, retired_at
  FROM enterprise_auth_bindings
 WHERE id = $1
   AND user_id = $2
 FOR UPDATE
`, authBindingID, userID)
	var binding enterpriseBindingRow
	if err := row.Scan(&binding.ID, &binding.UserID, &binding.ProviderID, &binding.ProviderKey, &binding.ProviderType, &binding.ProviderSubject, &binding.CreatedAt, &binding.LastAuthAt, &binding.RetiredAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return enterpriseBindingRow{}, ErrAuthBindingNotFound
		}
		return enterpriseBindingRow{}, err
	}
	return binding, nil
}

func hasActiveEnterpriseBindingForUserProviderTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, providerID uuid.UUID) bool {
	var exists bool
	_ = tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM enterprise_auth_bindings
     WHERE user_id = $1
       AND provider_id = $2
       AND retired_at IS NULL
)
`, userID, providerID).Scan(&exists)
	return exists
}

func hasActiveEnterpriseBindingForProviderSubjectTx(ctx context.Context, tx pgx.Tx, providerID uuid.UUID, providerSubject string) bool {
	var exists bool
	_ = tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM enterprise_auth_bindings
     WHERE provider_id = $1
       AND provider_subject = $2
       AND retired_at IS NULL
)
`, providerID, providerSubject).Scan(&exists)
	return exists
}

func decodeStringArrayJSON(raw []byte) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	if values == nil {
		return []string{}, nil
	}
	return values, nil
}

func encodeStringArrayJSON(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	data, err := json.Marshal(values)
	return string(data), err
}

func decodeOptionalSAMLSubjectSourceJSON(raw []byte) (*EnterpriseAuthSAMLSubjectSource, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var source EnterpriseAuthSAMLSubjectSource
	if err := json.Unmarshal(raw, &source); err != nil {
		return nil, err
	}
	return &source, nil
}

func encodeOptionalSAMLSubjectSourceJSON(source *EnterpriseAuthSAMLSubjectSource) (string, error) {
	if source == nil {
		return "null", nil
	}
	data, err := json.Marshal(source)
	return string(data), err
}

func updateUserVersionTx(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, userID uuid.UUID, changedAt time.Time) (UserRecord, error) {
	var updated UserRecord
	if err := tx.QueryRow(ctx, `
UPDATE users
   SET updated_at = $2,
       updated_by_user_id = $1,
       user_version = user_version + 1
 WHERE id = $3
RETURNING id, email::text, display_name, password_hash, password_changed_at, mfa_required, is_active, is_deployment_admin,
          created_at, updated_at, updated_by_user_id, last_login_at, user_version, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce
`, actorID, changedAt, userID).Scan(
		&updated.ID,
		&updated.Email,
		&updated.DisplayName,
		&updated.PasswordHash,
		&updated.PasswordChangedAt,
		&updated.MFARequired,
		&updated.IsActive,
		&updated.IsDeploymentAdmin,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.LastLoginAt,
		&updated.UserVersion,
		&updated.TOTPEnrolledAt,
		&updated.TOTPSecretCiphertext,
		&updated.TOTPSecretNonce,
	); err != nil {
		return UserRecord{}, err
	}
	return updated, nil
}

func nullableText(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
