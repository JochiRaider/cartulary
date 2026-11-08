package producer

import (
	"context"
	"encoding/json"
	"errors"
	stdflag "flag"
	"fmt"
	"io"
	"os"
	"strings"

	attackinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/attack"
	windowsinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/windows-event-ids"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// RunCommand composes one offline producer command with explicit arguments and
// output streams, returning its process exit status. Application runtime assembly
// must not import this package. Production always uses the admitted source auditor.
func RunCommand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return runCommand(ctx, args, stdout, stderr, Audit)
}

type attackAuditFunc func(context.Context, *os.Root, []byte, []byte, []byte, string, Builder) (AuditReport, error)

func runCommand(ctx context.Context, args []string, stdout, stderr io.Writer, audit attackAuditFunc) int {
	if err := executeCommand(ctx, args, stdout, stderr, audit); err != nil && !errors.Is(err, stdflag.ErrHelp) {
		fmt.Fprintln(stderr, "reference-pack-source:", err)
		return 1
	}
	return 0
}

func executeCommand(ctx context.Context, args []string, stdout, stderr io.Writer, audit attackAuditFunc) error {
	flags := stdflag.NewFlagSet("reference-pack-producer", stdflag.ContinueOnError)
	flags.SetOutput(stderr)
	action := flags.String("action", "check", "acquire, check, audit, identity")
	lockPath := flags.String("lock", "", "explicit source inventory or lock")
	directory := flags.String("directory", "", "explicit retained source root")
	builtAt := flags.String("built-at", "", "explicit reproducible UTC timestamp for audit")
	inputPath := flags.String("input", "", "closed package or envelope JSON input")
	editorialPath := flags.String("editorial", "", "optional Windows editorial JSON input")
	reportPath := flags.String("report-file", "", "optional source report destination")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected producer argument")
	}
	builder, err := BuilderIdentity()
	if err != nil {
		return err
	}
	if *action == "identity" {
		b, err := canonicaljson.Marshal(builder)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, string(b))
		return nil
	}
	if *action == "assemble" || *action == "envelope" || *action == "renew" {
		b, err := os.ReadFile(*inputPath)
		if err != nil {
			return err
		}
		var report any
		if *action == "assemble" {
			var header struct {
				SchemaID string `json:"schema_id"`
			}
			if err := json.Unmarshal(b, &header); err != nil {
				return err
			}
			if header.SchemaID == "cartulary.windows_event_build_input.v2" {
				report, err = AssembleWindows(ctx, b, builder)
			} else if header.SchemaID == "cartulary.veris_build_input.v1" {
				report, err = AssembleVERIS(ctx, b, builder)
			} else if header.SchemaID == "cartulary.entra_build_input.v2" {
				report, err = AssembleEntra(ctx, b, builder)
			} else if header.SchemaID == "cartulary.attack_package_input.v1" {
				report, err = Assemble(ctx, b, builder)
			} else {
				return errors.New("unsupported package input schema")
			}
		} else {
			report, err = Envelope(ctx, b, *action == "renew")
		}
		if err != nil {
			return err
		}
		encoded, err := canonicaljson.Marshal(report)
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, string(encoded))
		return nil
	}
	if *action != "discover" && *action != "acquire" && *action != "check" && *action != "audit" {
		return errors.New("unsupported source action")
	}
	if *directory == "" || !strings.HasSuffix(*lockPath, ".json") {
		return errors.New("producer inputs must be JSON, never prose")
	}
	lock, err := os.ReadFile(*lockPath)
	if err != nil {
		return err
	}
	if *action == "acquire" || *action == "discover" {
		if err := os.MkdirAll(*directory, 0700); err != nil {
			return err
		}
	}
	root, err := os.OpenRoot(*directory)
	if err != nil {
		return err
	}
	defer root.Close()
	var sourceHeader struct {
		SchemaID string `json:"schema_id"`
	}
	if err = json.Unmarshal(lock, &sourceHeader); err != nil {
		return err
	}
	if sourceHeader.SchemaID == "cartulary.veris_inventory.v1" {
		var report any
		switch *action {
		case "acquire":
			report, err = VERISAcquire(ctx, root, lock, VERISHTTPFetcher{})
		case "check":
			report, err = VERISCheck(root, lock)
		case "audit":
			report, err = VERISAuditSources(ctx, root, lock, builder)
		default:
			return errors.New("veris sources require the pinned inventory, not discovery")
		}
		if err != nil {
			return err
		}
		b, e := canonicaljson.Marshal(report)
		if e != nil {
			return e
		}
		if *reportPath != "" {
			if e = os.WriteFile(*reportPath, b, 0600); e != nil {
				return e
			}
		}
		fmt.Fprintf(stdout, "VERIS %s: report_sha256=%s; source_root=%s\n", *action, Digest(b), *directory)
		return nil
	}
	if sourceHeader.SchemaID == "cartulary.entra_inventory.v2" {
		var report any
		switch *action {
		case "acquire":
			report, err = EntraAcquire(ctx, root, lock, EntraHTTPFetcher{})
		case "check":
			report, err = EntraCheck(root, lock)
		case "audit":
			report, err = EntraAuditSources(ctx, root, lock, builder)
		default:
			return errors.New("entra sources require the pinned inventory, not discovery")
		}
		if err != nil {
			return err
		}
		b, e := canonicaljson.Marshal(report)
		if e != nil {
			return e
		}
		if *reportPath != "" {
			if e = os.WriteFile(*reportPath, b, 0600); e != nil {
				return e
			}
		}
		fmt.Fprintf(stdout, "Entra %s: report_sha256=%s; source_root=%s\n", *action, Digest(b), *directory)
		return nil
	}
	if sourceHeader.SchemaID == "cartulary.windows_event_inventory.v2" {
		var report any
		switch *action {
		case "discover":
			var inventory WindowsInventory
			inventory, err = WindowsDiscover(ctx, root, lock, WindowsHTTPFetcher{})
			if err == nil {
				b, e := canonicaljson.Marshal(inventory)
				if e != nil {
					return e
				}
				fmt.Fprintf(stdout, "Windows inventory: sources=%d; evidence=%s/inventories/%s.json\n", len(inventory.Sources), *directory, Digest(b))
			}
		case "acquire":
			var fetch WindowsFetcher = WindowsHTTPFetcher{}
			if *inputPath != "" {
				var local []byte
				local, err = os.ReadFile(*inputPath)
				if err == nil {
					fetch, err = WindowsLocalFetcher(lock, local, fetch)
				}
				if err != nil {
					return err
				}
			}
			report, err = WindowsAcquire(ctx, root, lock, fetch)
		case "check":
			report, err = WindowsCheck(root, lock)
		case "audit":
			var editorial []byte
			if *editorialPath != "" {
				editorial, err = os.ReadFile(*editorialPath)
			} else {
				editorial, err = windowsinputs.Files.ReadFile("editorial.v2.json")
			}
			if err != nil {
				return err
			}
			report, err = WindowsAudit(ctx, root, lock, editorial, builder)
		}
		if report != nil {
			b, e := canonicaljson.Marshal(report)
			if e != nil {
				return e
			}
			fmt.Fprintf(stdout, "Windows %s: report_sha256=%s; source_root=%s\n", *action, Digest(b), *directory)
			if *reportPath != "" {
				if e = os.WriteFile(*reportPath, b, 0600); e != nil {
					return e
				}
			}
		}
		return err
	}
	if *action == "discover" {
		return errors.New("discovery requires a Windows source inventory")
	}
	if sourceHeader.SchemaID != LockSchema {
		return errors.New("unsupported source inventory schema")
	}
	switch *action {
	case "acquire":
		_, err = Acquire(ctx, root, lock, HTTPFetcher{})
	case "check":
		_, err = Check(root, lock)
	case "audit":
		profile, readErr := attackinputs.Files.ReadFile("source-profile.native.v2.r2.json")
		if readErr != nil {
			return readErr
		}
		policy, readErr := attackinputs.Files.ReadFile("field-policy.native.v1.json")
		if readErr != nil {
			return readErr
		}
		report, auditErr := audit(ctx, root, lock, profile, policy, *builtAt, builder)
		return writeATTACKAuditReport(root, report, auditErr, *reportPath, *directory, stdout)
	}
	if err == nil {
		fmt.Fprintf(stdout, "source %s: verified pinned inputs; evidence=%s/acquisition.json\n", *action, *directory)
	}
	return err
}

func writeATTACKAuditReport(root *os.Root, report AuditReport, auditErr error, reportPath, directory string, stdout io.Writer) error {
	if report.SchemaID == "" {
		return auditErr
	}
	b, encodeErr := canonicaljson.Marshal(report)
	if encodeErr != nil {
		return encodeErr
	}
	name := "audits/" + Digest(b) + ".json"
	if err := Retain(root, name, b, 0600); err != nil {
		return err
	}
	if reportPath != "" {
		if err := os.WriteFile(reportPath, b, 0600); err != nil {
			return errors.Join(auditErr, err)
		}
	}
	fmt.Fprintf(stdout, "source audit: %s; records=%d; collisions=%d; issues=%d; evidence=%s/%s\n", report.Status, report.UniqueRecords, len(report.Collisions), len(report.Issues), directory, name)
	return auditErr
}
