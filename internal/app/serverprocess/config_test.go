package serverprocess

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func TestEffectiveConfigBackupRoot_Process(t *testing.T) {
	const websocketPath = "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v2/changes"

	validConfig := string(fixtures.MustRead("config", "valid.toml"))
	cases := []struct {
		name       string
		configText string
		mutateEnv  func(map[string]string)
		wantPath   string
		wantReason string
	}{
		{
			name:       "missing backup storage root",
			configText: stripConfigSection(t, validConfig, "[roots.backup_storage]"),
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__BINDING_KIND"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__SERVICE_REF"] = ""
			},
			wantPath:   "roots.backup_storage",
			wantReason: "missing_required_key",
		},
		{
			name:       "disconnected backup storage managed service",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__BINDING_KIND"] = "managed_service"
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__SERVICE_REF"] = "backup-vault"
			},
			wantPath:   "roots.backup_storage.binding_kind",
			wantReason: "profile_incompatible_binding",
		},
		{
			name:       "backup storage satisfied by export outputs",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = env["CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH"]
			},
			wantPath:   "roots.backup_storage.path",
			wantReason: "path_overlap",
		},
		{
			name:       "backup storage satisfied by temporary work",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = env["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"]
			},
			wantPath:   "roots.backup_storage.path",
			wantReason: "path_overlap",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := writeConfig(t, tc.configText)
			env := ConfigProcessEnv(t, configPath)
			tc.mutateEnv(env)

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected invalid backup-root config startup to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, websocketPath)
			server.RequireDiagnosticsCode(t, config.InvalidDeploymentConfigCode)
			server.RequireDiagnosticsField(t, tc.wantPath, tc.wantReason)
		})
	}
}

func ConfigProcessEnv(t testing.TB, configPath string) map[string]string {
	t.Helper()

	tempRoots := configtest.SetupTempRoots(t)
	env := map[string]string{
		config.ConfigFileEnv: configPath,
	}
	for key, value := range tempRoots.Paths {
		env[key] = value
	}
	configtest.EnsureRevisionsConflictTokenTestEnvironment(env)
	return env
}
