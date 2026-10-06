package operator

import (
	"context"
	"flag"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

// packagePreflight inspects configuration only. It never acquires a connection,
// initializes a root or substitutes package parsing for owner configuration.
func (runner operatorRunner) packagePreflight(_ context.Context, args []string) int {
	flags := flag.NewFlagSet("package preflight", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	source := flags.String("source-config-file", "", "source configuration")
	target := flags.String("target-config-file", "", "verification configuration")
	valid := flags.Parse(args[2:]) == nil && flags.NArg() == 0
	for _, path := range []string{*source, *target} {
		valid = valid && filepath.IsAbs(path) && filepath.Clean(path) == path && !strings.ContainsAny(path, "$\x00")
	}
	if !valid || *source == *target || runner.recovery.loadConfig == nil {
		return runner.packagePreflightResult(false)
	}
	sourceDeployment, err := runner.recovery.loadDeployment(*source)
	if err != nil {
		return runner.packagePreflightResult(false)
	}
	targetDeployment, err := runner.recovery.loadDeployment(*target)
	if err != nil {
		return runner.packagePreflightResult(false)
	}
	return runner.packagePreflightResult(packageBindingsValid(sourceDeployment, targetDeployment))
}

// packageReadiness is run inside the unpublished bootstrap container. It uses
// the same verified TLS policy as application clients and opens no listener.
func (runner operatorRunner) packageReadiness(ctx context.Context, args []string) int {
	if len(args) != 2 {
		return 2
	}
	tlsConfig, err := cryptography.TLSClient(cryptography.TLSClientOptions{
		ServerName: "localhost", RootCertificatePath: "/etc/cartulary/tls/ca.pem",
	})
	if err != nil {
		return 1
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://localhost:8080/readyz", nil)
	if err != nil {
		return 1
	}
	response, err := client.Do(request)
	if err != nil {
		return 1
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func packageBindingsValid(sourceDeployment, targetDeployment application.Deployment) bool {
	// This package provisions one PostgreSQL cluster and one object service.
	// Its fixed bindings deliberately do not admit arbitrary alias/provider maps.
	if sourceDeployment.DatabaseStorage.ServiceRef != "primary" || targetDeployment.DatabaseStorage.ServiceRef != "restore_verify" ||
		sourceDeployment.ObjectStorage.ServiceRef != "primary" || targetDeployment.ObjectStorage.ServiceRef != "restore_verify" {
		return false
	}
	sourceDB, sourceErr := pgx.ParseConfig(sourceDeployment.PostgresSettings.DSN)
	targetDB, targetErr := pgx.ParseConfig(targetDeployment.PostgresSettings.DSN)
	if sourceErr != nil || targetErr != nil || sourceDB.Host != "postgres" || targetDB.Host != sourceDB.Host ||
		sourceDB.Port != targetDB.Port || sourceDB.Database == targetDB.Database {
		return false
	}
	if sourceDeployment.ObjectSettings.Endpoint != "seaweedfs-s3:8333" ||
		targetDeployment.ObjectSettings.Endpoint != sourceDeployment.ObjectSettings.Endpoint ||
		sourceDeployment.ObjectSettings.Bucket == targetDeployment.ObjectSettings.Bucket {
		return false
	}
	for _, pair := range [][2]string{
		{sourceDeployment.BackupStorage.Path, "/var/lib/cartulary/backups"},
		{sourceDeployment.ReferencePackStorage.Path, "/var/lib/cartulary/reference-packs"},
		{sourceDeployment.ExportOutputs.Path, "/var/lib/cartulary/exports"},
		{targetDeployment.BackupStorage.Path, "/var/lib/cartulary/restore-verification-target/backups"},
		{targetDeployment.ReferencePackStorage.Path, "/var/lib/cartulary/restore-verification-target/reference-packs"},
		{targetDeployment.ExportOutputs.Path, "/var/lib/cartulary/restore-verification-target/exports"},
	} {
		if pair[0] != pair[1] {
			return false
		}
	}
	return true
}

func (runner operatorRunner) packagePreflightResult(ok bool) int {
	result, status := "failed", 2
	if ok {
		result, status = "succeeded", 0
	}
	if err := runner.recovery.transport.encodeJSON(struct {
		SchemaID string `json:"schema_id"`
		Result   string `json:"result"`
	}{"cartulary.package_preflight_result.v1", result}); err != nil {
		return 2
	}
	return status
}
