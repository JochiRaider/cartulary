package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/JochiRaider/cartulary/internal/platform/openapiassembly"
)

func main() {
	manifestPath := flag.String("manifest", "contracts/openapi-source/manifest.json", "owner-source manifest")
	check := flag.Bool("check", false, "compare assembled bytes with the target without writing")
	write := flag.Bool("write", false, "atomically replace the target when assembled bytes differ")
	flag.Parse()
	if *check == *write {
		fatal(errors.New("exactly one of --check or --write is required"))
	}
	output, target, err := openapiassembly.Assemble(*manifestPath)
	if err != nil {
		fatal(err)
	}
	if *check {
		if err := openapiassembly.CheckTarget(target, output); err != nil {
			fatal(err)
		}
		return
	}
	if err := openapiassembly.WriteTargetAtomically(target, output); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "openapi assembly failed: %v\n", err)
	os.Exit(1)
}
