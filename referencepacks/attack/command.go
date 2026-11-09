package producer

import (
	"context"
	"errors"
	stdflag "flag"
	"fmt"
	inputs "github.com/JochiRaider/cartulary/referencepacks/attack/inputs"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
	"io"
	"os"
)

// RunCommand composes this producer only. No application state or sibling
// producer participates in source acquisition, projection, or signing.
func RunCommand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	err := runCommand(ctx, args, stdout, stderr, Audit)
	if err != nil && !errors.Is(err, stdflag.ErrHelp) {
		fmt.Fprintln(stderr, "attack producer:", err)
		return 1
	}
	return 0
}

type commandAudit func(context.Context, *os.Root, []byte, []byte, []byte, string, Builder) (AuditReport, error)

func runCommand(ctx context.Context, args []string, stdout, stderr io.Writer, audit commandAudit) error {
	f := stdflag.NewFlagSet("attack", stdflag.ContinueOnError)
	f.SetOutput(stderr)
	action := f.String("action", "", "explicit source or publication action")
	lockPath := f.String("lock", "", "explicit pinned source lock or inventory")
	directory := f.String("directory", "", "retained research root")
	builtAt := f.String("built-at", "", "explicit UTC build timestamp")
	inputPath := f.String("input", "", "closed publication or local acquisition input")
	editorial := f.String("editorial", "", "explicit editorial input")
	reportPath := f.String("report-file", "", "optional report destination")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return errors.New("unexpected argument")
	}
	builder, err := BuilderIdentity()
	if err != nil {
		return err
	}
	var report any
	switch *action {
	case "identity":
		report = builder
	case "assemble", "assemble-fixture", "envelope", "renew", "envelope-fixture", "renew-fixture":
		raw, e := publication.ReadRegular(*inputPath, 4194304)
		if e != nil {
			return e
		}
		report, err = packageAction(ctx, *action, raw, builder)
		if err != nil {
			return err
		}
	case "acquire", "check", "audit", "discover":
		if *directory == "" {
			return errors.New("explicit source directory required")
		}
		lock, e := publication.ReadRegular(*lockPath, 4194304)
		if e != nil {
			return e
		}
		if *action == "acquire" || *action == "discover" {
			if e = os.MkdirAll(*directory, 0700); e != nil {
				return e
			}
		}
		root, e := os.OpenRoot(*directory)
		if e != nil {
			return e
		}
		defer root.Close()
		report, err = sourceAction(ctx, *action, root, lock, *builtAt, *editorial, *inputPath, builder, audit)
		if err != nil && *action != "audit" {
			return err
		}
	default:
		return errors.New("unsupported explicit producer action")
	}
	if *action == "audit" && err == nil {
		if err = validateAuditReport(report); err != nil {
			return err
		}
	}
	if report != nil {
		raw, e := canonicaljson.Marshal(report)
		if e != nil {
			return errors.Join(err, e)
		}
		if *reportPath != "" {
			if e = os.WriteFile(*reportPath, raw, 0600); e != nil {
				return errors.Join(err, e)
			}
		}
		if _, e = fmt.Fprintln(stdout, string(raw)); e != nil {
			return errors.Join(err, e)
		}
	}
	return err
}
func packageAction(ctx context.Context, action string, input []byte, builder Builder) (any, error) {
	switch action {
	case "assemble":
		return AssembleRelease(ctx, input, builder)
	case "assemble-fixture":
		return AssembleFixture(ctx, input, builder)
	case "envelope", "renew":
		var in ReleaseEnvelopeInput
		if err := DecodeClosed(input, &in); err != nil {
			return nil, err
		}
		if (action == "renew") != (in.Envelope.PreviousEnvelope != "") {
			return nil, errors.New("renewal requires an explicit predecessor; first publication forbids one")
		}
		return PublishRelease(ctx, input, publication.SystemClock{})
	case "envelope-fixture", "renew-fixture":
		var in FixtureEnvelopeInput
		if err := DecodeClosed(input, &in); err != nil {
			return nil, err
		}
		if (action == "renew-fixture") != (in.Envelope.PreviousEnvelope != "") {
			return nil, errors.New("fixture renewal predecessor mismatch")
		}
		return PublishFixture(ctx, input, publication.SystemClock{})
	default:
		return nil, errors.New("unsupported publication action")
	}
}
func sourceAction(ctx context.Context, action string, root *os.Root, lock []byte, builtAt, editorial, sourceInput string, builder Builder, audit commandAudit) (any, error) {
	switch action {
	case "acquire":
		return Acquire(ctx, root, lock, HTTPFetcher{})
	case "check":
		return Check(root, lock)
	case "audit":
		profile, err := inputs.Files.ReadFile("dataset-profile.v2.json")
		if err != nil {
			return nil, err
		}
		policy, err := inputs.Files.ReadFile("field-policy.native.v1.json")
		if err != nil {
			return nil, err
		}
		report, err := audit(ctx, root, lock, profile, policy, builtAt, builder)
		if report.SchemaID == "" {
			return nil, err
		}
		{
			raw, e := canonicaljson.Marshal(report)
			if e != nil {
				return nil, e
			}
			if e = Retain(root, "audits/"+Digest(raw)+".json", raw, 0600); e != nil {
				return nil, e
			}
		}
		return report, err
	}
	return nil, errors.New("unsupported source action")
}
