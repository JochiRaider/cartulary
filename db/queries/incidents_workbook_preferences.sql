-- name: GetDefaultWorkbookPreferences :one
SELECT
    incident_id,
    default_sheet_ref,
    created_at,
    updated_at,
    updated_by_user_id
FROM incident_workbook_preferences
WHERE incident_id = $1;

-- name: PutDefaultWorkbookPreferences :one
INSERT INTO incident_workbook_preferences (
    incident_id,
    default_sheet_ref,
    created_at,
    updated_at,
    updated_by_user_id
)
VALUES ($1, $2::jsonb, $3, $3, $4)
ON CONFLICT (incident_id) DO UPDATE
SET
    default_sheet_ref = EXCLUDED.default_sheet_ref,
    updated_at = CASE
        WHEN incident_workbook_preferences.default_sheet_ref IS NOT DISTINCT FROM EXCLUDED.default_sheet_ref
        THEN incident_workbook_preferences.updated_at
        ELSE EXCLUDED.updated_at
    END,
    updated_by_user_id = CASE
        WHEN incident_workbook_preferences.default_sheet_ref IS NOT DISTINCT FROM EXCLUDED.default_sheet_ref
        THEN incident_workbook_preferences.updated_by_user_id
        ELSE EXCLUDED.updated_by_user_id
    END
RETURNING
    incident_id,
    default_sheet_ref,
    created_at,
    updated_at,
    updated_by_user_id;

-- name: GetUserWorkbookPreferences :one
SELECT
    incident_id,
    user_id,
    home_sheet_ref,
    created_at,
    updated_at
FROM user_workbook_preferences
WHERE incident_id = $1
  AND user_id = $2;

-- name: PutUserWorkbookPreferences :one
INSERT INTO user_workbook_preferences (
    incident_id,
    user_id,
    home_sheet_ref,
    created_at,
    updated_at
)
VALUES ($1, $2, $3::jsonb, $4, $4)
ON CONFLICT (incident_id, user_id) DO UPDATE
SET
    home_sheet_ref = EXCLUDED.home_sheet_ref,
    updated_at = CASE
        WHEN user_workbook_preferences.home_sheet_ref IS NOT DISTINCT FROM EXCLUDED.home_sheet_ref
        THEN user_workbook_preferences.updated_at
        ELSE EXCLUDED.updated_at
    END
RETURNING
    incident_id,
    user_id,
    home_sheet_ref,
    created_at,
    updated_at;

-- name: GetUserWorkbookPreferenceRefForUpdate :one
SELECT home_sheet_ref
FROM user_workbook_preferences
WHERE incident_id = $1
  AND user_id = $2
FOR UPDATE;

-- name: GetDefaultWorkbookPreferenceRefForUpdate :one
SELECT default_sheet_ref
FROM incident_workbook_preferences
WHERE incident_id = $1
FOR UPDATE;

-- name: ClearUserWorkbookPreferenceRef :execrows
UPDATE user_workbook_preferences
SET
    home_sheet_ref = NULL,
    updated_at = $4
WHERE incident_id = $1
  AND user_id = $2
  AND home_sheet_ref IS NOT DISTINCT FROM $3::jsonb;

-- name: ClearDefaultWorkbookPreferenceRef :execrows
UPDATE incident_workbook_preferences
SET
    default_sheet_ref = NULL,
    updated_at = $4,
    updated_by_user_id = $3
WHERE incident_id = $1
  AND default_sheet_ref IS NOT DISTINCT FROM $2::jsonb;

-- name: InsertIncidentWorkbookPreferencesBootstrap :exec
INSERT INTO incident_workbook_preferences (
    incident_id,
    default_sheet_ref,
    created_at,
    updated_at,
    updated_by_user_id
)
VALUES ($1, NULL, $2, $2, $3);

-- name: InsertUserWorkbookPreferencesBootstrap :exec
INSERT INTO user_workbook_preferences (
    incident_id,
    user_id,
    home_sheet_ref,
    created_at,
    updated_at
)
VALUES ($1, $2, NULL, $3, $3);
