package config

import "testing"

func TestBackupStorageRootBindingConfig(t *testing.T) {
	t.Run("requires backup storage for every supported deployment profile", func(t *testing.T) {
		for _, profile := range backupSupportedDeploymentProfiles() {
			t.Run(profile, func(t *testing.T) {
				cfg := backupDeploymentProfileConfig(t, profile)
				cfg.Roots.BackupStorage = RootBinding{}

				_, err := validate(cfg)
				requireDiagnostic(t, err, "roots.backup_storage", "missing_required_key")
			})
		}
	})

	t.Run("accepts filesystem root backup storage for every supported deployment profile", func(t *testing.T) {
		for _, profile := range backupSupportedDeploymentProfiles() {
			t.Run(profile, func(t *testing.T) {
				if _, err := validate(backupDeploymentProfileConfig(t, profile)); err != nil {
					t.Fatalf("validate %s filesystem-root backup storage: %v", profile, err)
				}
			})
		}
	})

	t.Run("accepts managed service backup storage only outside disconnected profile", func(t *testing.T) {
		for _, profile := range backupSupportedDeploymentProfiles() {
			t.Run(profile, func(t *testing.T) {
				cfg := backupDeploymentProfileConfig(t, profile)
				cfg.Roots.BackupStorage = RootBinding{
					BindingKind: "managed_service",
					ServiceRef:  "backup-vault",
				}

				_, err := validate(cfg)
				if profile != "disconnected" {
					if err != nil {
						t.Fatalf("validate %s managed-service backup storage: %v", profile, err)
					}
					return
				}

				requireDiagnostic(t, err, "roots.backup_storage.binding_kind", "profile_incompatible_binding")
			})
		}
	})

	t.Run("rejects unknown backup storage binding kind", func(t *testing.T) {
		cfg := backupDeploymentProfileConfig(t, "on_prem")
		cfg.Roots.BackupStorage = RootBinding{
			BindingKind: "object_store",
			ServiceRef:  "backup-vault",
		}

		_, err := validate(cfg)
		requireDiagnostic(t, err, "roots.backup_storage.binding_kind", "invalid_enum")
	})

	t.Run("rejects export and temporary-work roots as backup storage", func(t *testing.T) {
		cases := []struct {
			name string
			path string
		}{
			{name: "export_outputs", path: backupDeploymentProfileConfig(t, "disconnected").Roots.ExportOutputs.Path},
			{name: "temporary_work", path: backupDeploymentProfileConfig(t, "disconnected").Roots.TemporaryWork.Path},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := backupDeploymentProfileConfig(t, "disconnected")
				cfg.Roots.BackupStorage.Path = tc.path

				_, err := validate(cfg)
				requireDiagnostic(t, err, "roots.backup_storage.path", "path_overlap")
			})
		}
	})
}

func backupDeploymentProfileConfig(t testing.TB, profile string) document {
	t.Helper()

	return bootstrapDeploymentProfileConfig(t, profile)
}

func backupSupportedDeploymentProfiles() []string {
	return []string{"disconnected", "on_prem", "cloud"}
}
