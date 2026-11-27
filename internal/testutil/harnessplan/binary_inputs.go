package harnessplan

import (
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"
)

const producerRegistryPath = "tools/harness_producers.json"

// binaryProducer is the part of a producer declaration a closure reads.
type binaryProducer struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Binding       string   `json:"binding"`
	Directory     string   `json:"directory"`
	Target        string   `json:"target"`
	Package       string   `json:"package"`
	Source        string   `json:"source"`
	ClientSupport string   `json:"client_support"`
	Inputs        []string `json:"inputs"`
}

// binaryInputs adds what each runtime binary a unit's rows run is built from
// (P6-6). A row that runs a built server, operator or migrate binary depends
// on that binary as much as on its own test code, so the binary's Go import
// closure without tests joins the unit's closure. Where that closure embeds a
// generated directory, such as the web assets, the declared inputs it is
// generated from join too. The producer registry and the workspace layout
// say both; nothing is listed by hand. It returns the binaries' package
// directories.
func (p *cachePlanner) binaryInputs(ids []string, set map[int]bool) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var registry struct {
		Producers []binaryProducer `json:"producers"`
	}
	var layout struct {
		Directories []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"directories"`
	}
	for file, into := range map[string]any{producerRegistryPath: &registry, workspaceLayoutPath: &layout} {
		text, err := p.text(file)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(text), into); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		p.under(file, set)
	}
	var packages []string
	for _, id := range ids {
		index := slices.IndexFunc(p.in.Topology.RuntimeBinaries, func(binary RuntimeBinary) bool { return binary.ID == id })
		if index < 0 {
			return nil, fmt.Errorf("runtime binary %s is not in the topology", id)
		}
		binding := p.in.Topology.RuntimeBinaries[index].OutputMakeVariable
		producer := slices.IndexFunc(registry.Producers, func(d binaryProducer) bool { return d.Kind == "go_binary" && d.Binding == binding })
		if producer < 0 || !strings.HasPrefix(registry.Producers[producer].Package, "./") {
			return nil, fmt.Errorf("runtime binary %s has no Go binary producer", id)
		}
		packages = append(packages, strings.TrimSuffix(strings.TrimPrefix(registry.Producers[producer].Package, "./"), "/"))
	}
	directories, err := p.packageDirectories(dedupe(packages), false)
	if err != nil {
		return nil, err
	}
	for _, directory := range directories {
		p.withoutTests(directory, set)
		// The nearest module's files, the repository root's included, fix
		// the dependency versions the binary is built with.
		for module := directory; ; module = path.Dir(module) {
			if _, ok := p.index(path.Join(module, "go.mod")); ok {
				p.under(path.Join(module, "go.mod"), set)
				p.under(path.Join(module, "go.sum"), set)
				break
			}
			if module == "." {
				break
			}
		}
	}
	layoutPaths := map[string]string{}
	for _, entry := range layout.Directories {
		layoutPaths[entry.ID] = entry.Path
	}
	for _, embedded := range registry.Producers {
		generated := layoutPaths[embedded.Directory]
		if embedded.Kind != "embedded_assets" || generated == "" ||
			!slices.ContainsFunc(directories, func(directory string) bool { return strings.HasPrefix(generated, directory+"/") }) {
			continue
		}
		source := slices.IndexFunc(registry.Producers, func(d binaryProducer) bool { return d.Kind == "web_bundle" && d.Target == embedded.Source })
		if source < 0 {
			return nil, fmt.Errorf("embedded assets %s name no web bundle %s", embedded.ID, embedded.Source)
		}
		for _, input := range append(slices.Clone(registry.Producers[source].Inputs), embedded.ClientSupport) {
			if input != "" {
				p.under(input, set)
			}
		}
	}
	return directories, nil
}
