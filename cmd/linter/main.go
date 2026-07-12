// Command linter is a multichecker assembled from several static
// analyzers:
//
//   - standard analyzers from golang.org/x/tools/go/analysis/passes;
//   - all SA-class analyzers from staticcheck (honnef.co/go/tools);
//   - selected analyzers from the other staticcheck classes (S1000, ST1000);
//   - public analyzers errcheck and bodyclose;
//   - the own exitcheck analyzer, which forbids a direct os.Exit call in the
//     main function of package main.
//
// The set of enabled staticcheck-family analyzers is driven by config.json,
// which is embedded into the binary. Each entry is matched against an analyzer
// name either exactly or, when it ends with "*", as a prefix (e.g. "SA*"
// enables every SA-class check).
//
// Usage:
//
//	linter [flags] package...
//
// For example, to run all configured analyzers over the whole module:
//
//	go run ./cmd/linter ./...
package main

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"Ustasjs/yp-url-shortener/cmd/linter/exitcheck"
)

// config holds the embedded multichecker configuration.
//
//go:embed config.json
var configData []byte

// Config describes which staticcheck-family analyzers are enabled.
type Config struct {
	Staticcheck []string `json:"staticcheck"`
}

func main() {
	cfg := mustParseConfig(configData)

	analyzers := standardPasses()
	analyzers = append(analyzers, staticcheckAnalyzers(cfg)...)
	analyzers = append(analyzers,
		errcheck.Analyzer,
		bodyclose.Analyzer,
		exitcheck.Analyzer,
	)

	multichecker.Main(analyzers...)
}

// mustParseConfig parses the embedded configuration, panicking on error since
// the data is baked into the binary at build time.
func mustParseConfig(data []byte) Config {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic("linter: invalid config.json: " + err.Error())
	}
	return cfg
}

// standardPasses returns the standard analyzers from
// golang.org/x/tools/go/analysis/passes.
func standardPasses() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
	}
}

// staticcheckAnalyzers returns the staticcheck, simple and stylecheck
// analyzers whose names are enabled by the configuration.
func staticcheckAnalyzers(cfg Config) []*analysis.Analyzer {
	var result []*analysis.Analyzer
	for _, group := range [][]*lint.Analyzer{
		staticcheck.Analyzers,
		simple.Analyzers,
		stylecheck.Analyzers,
	} {
		for _, a := range group {
			if cfg.enabled(a.Analyzer.Name) {
				result = append(result, a.Analyzer)
			}
		}
	}
	return result
}

// enabled reports whether the analyzer name matches any configured entry,
// either exactly or as a "prefix*" pattern.
func (c Config) enabled(name string) bool {
	for _, entry := range c.Staticcheck {
		if prefix, ok := strings.CutSuffix(entry, "*"); ok {
			if strings.HasPrefix(name, prefix) {
				return true
			}
		} else if entry == name {
			return true
		}
	}
	return false
}
