// The producer entrypoint composes retained source and isolated fixture package
// production. Runtime admission and deployment authority remain separate.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	windowsinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/windows-event-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func run() error {
	action := flag.String("action", "check", "acquire, check, audit, identity")
	lockPath := flag.String("lock", "", "explicit source inventory or lock")
	candidatePath := flag.String("candidate", "contracts/reference-pack-producers/attack/candidate-index.v1.json", "producer-only candidate index")
	directory := flag.String("directory", "", "explicit retained source root")
	builtAt := flag.String("built-at", "", "explicit reproducible UTC timestamp for audit")
	inputPath := flag.String("input", "", "closed package or envelope JSON input")
	editorialPath := flag.String("editorial", "", "optional Windows editorial JSON input")
	reportPath := flag.String("report-file", "", "optional source report destination")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected producer argument")
	}
	builder, err := producer.BuilderIdentity()
	if err != nil {
		return err
	}
	if *action == "identity" {
		b, err := canonicaljson.Marshal(builder)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
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
				report, err = producer.AssembleWindows(context.Background(), b, builder)
			} else if header.SchemaID == "cartulary.base_registry_reissue_input.v1" {
				report, err = producer.ReissueBase(context.Background(), b, builder)
			} else {
				report, err = producer.Assemble(context.Background(), b, builder)
			}
		} else {
			report, err = producer.Envelope(context.Background(), b, *action == "renew")
		}
		if err != nil {
			return err
		}
		encoded, err := canonicaljson.Marshal(report)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}
	if *action != "discover" && *action != "acquire" && *action != "check" && *action != "audit" {
		return errors.New("unsupported source action")
	}
	if *directory == "" || !strings.HasSuffix(*lockPath, ".json") || !strings.HasSuffix(*candidatePath, ".json") {
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
	if sourceHeader.SchemaID == "cartulary.windows_event_inventory.v2" {
		var report any
		switch *action {
		case "discover":
			var inventory producer.WindowsInventory
			inventory, err = producer.WindowsDiscover(context.Background(), root, lock, producer.WindowsHTTPFetcher{})
			if err == nil {
				b, e := canonicaljson.Marshal(inventory)
				if e != nil {
					return e
				}
				fmt.Printf("Windows inventory: sources=%d; evidence=%s/inventories/%s.json\n", len(inventory.Sources), *directory, producer.Digest(b))
			}
		case "acquire":
			var fetch producer.WindowsFetcher = producer.WindowsHTTPFetcher{}
			if *inputPath != "" {
				var local []byte
				local, err = os.ReadFile(*inputPath)
				if err == nil {
					fetch, err = producer.WindowsLocalFetcher(lock, local, fetch)
				}
				if err != nil {
					return err
				}
			}
			report, err = producer.WindowsAcquire(context.Background(), root, lock, fetch)
		case "check":
			report, err = producer.WindowsCheck(root, lock)
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
			report, err = producer.WindowsAudit(context.Background(), root, lock, editorial, builder)
		}
		if report != nil {
			b, e := canonicaljson.Marshal(report)
			if e != nil {
				return e
			}
			fmt.Printf("Windows %s: report_sha256=%s; source_root=%s\n", *action, producer.Digest(b), *directory)
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
	switch *action {
	case "acquire":
		_, err = producer.Acquire(context.Background(), root, lock, producer.HTTPFetcher{})
	case "check":
		_, err = producer.Check(root, lock)
	case "audit":
		indexBytes, readErr := os.ReadFile(*candidatePath)
		if readErr != nil {
			return readErr
		}
		candidateRoot, openErr := os.OpenRoot(filepath.Dir(*candidatePath))
		if openErr != nil {
			return openErr
		}
		defer candidateRoot.Close()
		files, readErr := producer.ReadCandidate(candidateRoot, indexBytes)
		if readErr != nil {
			return readErr
		}
		retainedPrefix := "candidates/" + producer.Digest(indexBytes) + "/"
		for name, b := range files {
			if err := producer.Retain(root, retainedPrefix+name, b, 0600); err != nil {
				return err
			}
		}
		if err := producer.Retain(root, retainedPrefix+"candidate-index.v1.json", indexBytes, 0600); err != nil {
			return err
		}
		policy := files["field-policy.candidate.v1.json"]
		report, auditErr := producer.Audit(context.Background(), root, lock, policy, *builtAt, builder)
		if report.SchemaID == "" {
			return auditErr
		}
		if auditErr == nil {
			native, nativeErr := producer.ValidateNativeCandidate(context.Background(), root, lock, indexBytes, files["native-source.candidate.v1.schema.json"])
			report.NativeCandidate = &native
			if nativeErr != nil {
				report.Status = "source_audit_failed"
				report.Issues = append(report.Issues, "native_candidate_failed")
				auditErr = nativeErr
			}
		}
		b, encodeErr := canonicaljson.Marshal(report)
		if encodeErr != nil {
			return encodeErr
		}
		name := "audits/" + producer.Digest(b) + ".json"
		if err := producer.Retain(root, name, b, 0600); err != nil {
			return err
		}
		fmt.Printf("source audit: %s; records=%d; collisions=%d; issues=%d; evidence=%s/%s\n", report.Status, report.UniqueRecords, len(report.Collisions), len(report.Issues), *directory, name)
		return auditErr
	}
	if err == nil {
		fmt.Printf("source %s: verified pinned inputs; evidence=%s/acquisition.json\n", *action, *directory)
	}
	return err
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "reference-pack-source:", err)
		os.Exit(1)
	}
}
