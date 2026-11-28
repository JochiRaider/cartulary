// Command gosecdriver runs gosec's rules and analyzers with one package load
// per module. The gosec command loads each package separately, and every load
// walks the whole dependency graph again with its own `go list -deps -export`;
// for this repository that repetition costs about thirty times the analysis
// (Testing Harness G67). The driver keeps the command's meaning: the same
// package discovery, build-constraint filtering, rule and analyzer selection,
// suppression handling, report and exit status, for the flags the
// static-analysis wrappers use. Any other flag is refused.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"sync"

	"github.com/securego/gosec/v2"
	"github.com/securego/gosec/v2/analyzers"
	"github.com/securego/gosec/v2/issue"
	"github.com/securego/gosec/v2/report"
	"github.com/securego/gosec/v2/rules"
	"golang.org/x/tools/go/packages"
)

const (
	exitSuccess = 0
	exitFailure = 1
	gosecModule = "github.com/securego/gosec/v2"
)

type directories []string

func (d *directories) String() string { return strings.Join(*d, " ") }

func (d *directories) Set(value string) error {
	*d = append(*d, value)
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type options struct {
	include          string
	excludeGenerated bool
	noFail           bool
	quiet            bool
	terse            bool
	concurrency      int
	format           string
	excludeDirs      directories
	patterns         []string
}

func parse(args []string, stderr io.Writer) (options, error) {
	o := options{excludeDirs: directories{"vendor", `\.git/`}}
	flags := flag.NewFlagSet("gosec", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&o.include, "include", "", "Comma separated list of rules IDs to include")
	flags.BoolVar(&o.excludeGenerated, "exclude-generated", false, "Exclude generated files")
	flags.BoolVar(&o.noFail, "no-fail", false, "Do not fail the scanning, even if issues were found")
	flags.BoolVar(&o.quiet, "quiet", false, "Only show output when errors are found")
	flags.BoolVar(&o.terse, "terse", false, "Shows only the results and summary")
	flags.IntVar(&o.concurrency, "concurrency", runtime.GOMAXPROCS(0), "Packages analyzed at once")
	flags.StringVar(&o.format, "fmt", "text", "Report format, as gosec's -fmt")
	flags.Var(&o.excludeDirs, "exclude-dir", "Exclude folder from scan (can be specified multiple times)")
	if err := flags.Parse(args); err != nil {
		return o, err
	}
	if flags.NArg() == 0 {
		return o, errors.New("FILE [FILE...] or './...' expected")
	}
	if o.concurrency < 1 {
		return o, errors.New("-concurrency must be positive")
	}
	o.patterns = flags.Args()
	return o, nil
}

func run(args []string, stdout, stderr io.Writer) int {
	o, err := parse(args, stderr)
	if err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(stderr, "gosec: %v\n", err)
		}
		return exitFailure
	}
	logger := log.New(stderr, "[gosec] ", log.LstdFlags)
	if o.quiet || o.terse {
		logger = log.New(io.Discard, "", 0)
	}

	config := gosec.NewConfig()
	config.SetGlobal(gosec.IncludeRules, o.include)
	config.SetGlobal(gosec.ExcludeRules, "")
	var ruleFilters []rules.RuleFilter
	var analyzerFilters []analyzers.AnalyzerFilter
	if o.include != "" {
		ids := strings.Split(o.include, ",")
		ruleFilters = append(ruleFilters, rules.NewRuleFilter(false, ids...))
		analyzerFilters = append(analyzerFilters, analyzers.NewAnalyzerFilter(false, ids...))
	}
	ruleList := rules.Generate(false, ruleFilters...)
	analyzerList := analyzers.Generate(false, analyzerFilters...)
	if len(ruleList.Rules) == 0 && len(analyzerList.Analyzers) == 0 {
		logger.Print("No rules/analyzers are configured")
		return exitFailure
	}

	excluded := gosec.ExcludedDirsRegExp(o.excludeDirs)
	var dirs []string
	for _, pattern := range o.patterns {
		found, err := gosec.PackagePaths(pattern, excluded)
		if err != nil {
			logger.Printf("Failed to get package paths: %v", err)
			return exitFailure
		}
		dirs = append(dirs, found...)
	}
	slices.Sort(dirs)
	dirs = slices.Compact(dirs)
	if len(dirs) == 0 {
		logger.Print("No packages found")
		return exitFailure
	}

	scan := newScan(config, ruleList, analyzerList, o, logger)
	modules, err := scan.group(dirs)
	if err != nil {
		logger.Printf("Analyzer error: %v", err)
		return exitFailure
	}
	for _, module := range modules {
		if err := scan.analyze(module); err != nil {
			logger.Printf("Analyzer error: %v", err)
			return exitFailure
		}
	}

	issues, metrics, fileErrors := scan.report()
	sortIssues(issues)
	trueIssues := 0
	for _, found := range issues {
		if len(found.Suppressions) == 0 {
			trueIssues++
		}
	}
	metrics.NumFound = trueIssues
	if len(issues) == 0 && o.quiet {
		return exitSuccess
	}
	rootPaths := make([]string, 0, len(o.patterns))
	for _, pattern := range o.patterns {
		rootPath, err := gosec.RootPath(pattern)
		if err != nil {
			logger.Printf("Failed to get root paths: %v", err)
			return exitFailure
		}
		rootPaths = append(rootPaths, rootPath)
	}
	reportInfo := gosec.NewReportInfo(issues, metrics, fileErrors).WithVersion(version())
	if err := report.CreateReport(stdout, o.format, true, rootPaths, reportInfo); err != nil {
		logger.Printf("Failed to print report: %v", err)
		return exitFailure
	}
	if (trueIssues > 0 || len(fileErrors) > 0) && !o.noFail {
		return exitFailure
	}
	return exitSuccess
}

// module is one module root and the package directories under it to load.
type module struct {
	root string
	dirs []string
}

// scan accumulates one run's findings across modules and its workers.
type scan struct {
	config       gosec.Config
	ruleList     rules.RuleList
	analyzerList *analyzers.AnalyzerList
	options      options
	logger       *log.Logger

	mu      sync.Mutex
	issues  []*issue.Issue
	metrics *gosec.Metrics
	errors  map[string][]gosec.Error
}

func newScan(config gosec.Config, ruleList rules.RuleList, analyzerList *analyzers.AnalyzerList, o options, logger *log.Logger) *scan {
	return &scan{config: config, ruleList: ruleList, analyzerList: analyzerList, options: o, logger: logger,
		metrics: &gosec.Metrics{}, errors: map[string][]gosec.Error{}}
}

func (s *scan) appendError(file string, err error) {
	s.errors[file] = append(s.errors[file], *gosec.NewError(0, 0, err.Error()))
}

// group applies gosec's per-directory build-constraint filter, then groups
// the buildable directories by module root, which each load runs in. A
// directory with no buildable Go files is skipped, as gosec skips it.
func (s *scan) group(dirs []string) ([]module, error) {
	context := build.Default
	byRoot := map[string]*module{}
	var roots []string
	for _, dir := range dirs {
		if _, err := context.ImportDir(dir, build.ImportComment); err != nil {
			var noGo *build.NoGoError
			if !errors.As(err, &noGo) {
				s.appendError(dir, fmt.Errorf("importing dir %q: %w", dir, err))
			}
			continue
		}
		absolute, err := gosec.GetPkgAbsPath(dir)
		if err != nil {
			s.logger.Printf("Skipping: %s. Path doesn't exist.", absolute)
			continue
		}
		root := gosec.FindModuleRoot(absolute)
		if root == "" {
			return nil, fmt.Errorf("%s is in no module", dir)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			return nil, err
		}
		if byRoot[root] == nil {
			byRoot[root] = &module{root: root}
			roots = append(roots, root)
		}
		byRoot[root].dirs = append(byRoot[root].dirs, "./"+filepath.ToSlash(relative))
	}
	slices.Sort(roots)
	modules := make([]module, 0, len(roots))
	for _, root := range roots {
		modules = append(modules, *byRoot[root])
	}
	return modules, nil
}

// analyze loads a module's packages once and checks them on workers, each
// with its own analyzer, since an analyzer's checks share its state.
func (s *scan) analyze(m module) error {
	loaded, err := packages.Load(&packages.Config{Mode: gosec.LoadMode, Dir: m.root}, m.dirs...)
	if err != nil {
		return fmt.Errorf("loading packages of %s: %w", m.root, err)
	}
	jobs := make(chan *packages.Package)
	var workers sync.WaitGroup
	for range min(s.options.concurrency, max(len(loaded), 1)) {
		workers.Go(func() {
			analyzer := gosec.NewAnalyzer(s.config, false, s.options.excludeGenerated, false, 1, s.logger)
			analyzer.LoadRules(s.ruleList.RulesInfo())
			analyzer.LoadAnalyzers(s.analyzerList.AnalyzersInfo())
			parseErrors := map[string][]gosec.Error{}
			for pkg := range jobs {
				if pkg.Name == "" {
					continue
				}
				found, err := gosec.ParseErrors(pkg)
				if err != nil {
					s.mu.Lock()
					s.appendError(pkg.PkgPath, fmt.Errorf("parsing errors in pkg %q: %w", pkg.Name, err))
					s.mu.Unlock()
					continue
				}
				for file, errs := range found {
					parseErrors[file] = append(parseErrors[file], errs...)
				}
				analyzer.CheckRules(pkg)
				analyzer.CheckAnalyzers(pkg)
			}
			issues, metrics, analyzerErrors := analyzer.Report()
			s.mu.Lock()
			defer s.mu.Unlock()
			s.issues = append(s.issues, issues...)
			s.metrics.Merge(metrics)
			for _, source := range []map[string][]gosec.Error{parseErrors, analyzerErrors} {
				for file, errs := range source {
					s.errors[file] = append(s.errors[file], errs...)
				}
			}
		})
	}
	for _, pkg := range loaded {
		jobs <- pkg
	}
	close(jobs)
	workers.Wait()
	return nil
}

func (s *scan) report() ([]*issue.Issue, *gosec.Metrics, map[string][]gosec.Error) {
	for file := range s.errors {
		slices.SortFunc(s.errors[file], func(a, b gosec.Error) int {
			if a.Line != b.Line {
				return a.Line - b.Line
			}
			return a.Column - b.Column
		})
	}
	return s.issues, s.metrics, s.errors
}

// version is the gosec module version the driver is built with.
func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dependency := range info.Deps {
			if dependency.Path == gosecModule {
				return dependency.Version
			}
		}
	}
	return "dev"
}
