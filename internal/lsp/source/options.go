// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package source

import (
	"fmt"
	"os"
	"time"

	"github.com/jackie-feng/tools/go/analysis"
	"github.com/jackie-feng/tools/go/analysis/passes/asmdecl"
	"github.com/jackie-feng/tools/go/analysis/passes/assign"
	"github.com/jackie-feng/tools/go/analysis/passes/atomic"
	"github.com/jackie-feng/tools/go/analysis/passes/atomicalign"
	"github.com/jackie-feng/tools/go/analysis/passes/bools"
	"github.com/jackie-feng/tools/go/analysis/passes/buildtag"
	"github.com/jackie-feng/tools/go/analysis/passes/cgocall"
	"github.com/jackie-feng/tools/go/analysis/passes/composite"
	"github.com/jackie-feng/tools/go/analysis/passes/copylock"
	"github.com/jackie-feng/tools/go/analysis/passes/httpresponse"
	"github.com/jackie-feng/tools/go/analysis/passes/loopclosure"
	"github.com/jackie-feng/tools/go/analysis/passes/lostcancel"
	"github.com/jackie-feng/tools/go/analysis/passes/nilfunc"
	"github.com/jackie-feng/tools/go/analysis/passes/printf"
	"github.com/jackie-feng/tools/go/analysis/passes/shift"
	"github.com/jackie-feng/tools/go/analysis/passes/sortslice"
	"github.com/jackie-feng/tools/go/analysis/passes/stdmethods"
	"github.com/jackie-feng/tools/go/analysis/passes/structtag"
	"github.com/jackie-feng/tools/go/analysis/passes/tests"
	"github.com/jackie-feng/tools/go/analysis/passes/unmarshal"
	"github.com/jackie-feng/tools/go/analysis/passes/unreachable"
	"github.com/jackie-feng/tools/go/analysis/passes/unsafeptr"
	"github.com/jackie-feng/tools/go/analysis/passes/unusedresult"
	"github.com/jackie-feng/tools/internal/lsp/diff"
	"github.com/jackie-feng/tools/internal/lsp/diff/myers"
	"github.com/jackie-feng/tools/internal/lsp/protocol"
	"github.com/jackie-feng/tools/internal/telemetry/tag"
	errors "golang.org/x/xerrors"
)

var (
	DefaultOptions = Options{
		Env:                    os.Environ(),
		TextDocumentSyncKind:   protocol.Incremental,
		HoverKind:              SynopsisDocumentation,
		InsertTextFormat:       protocol.PlainTextTextFormat,
		PreferredContentFormat: protocol.Markdown,
		SupportedCodeActions: map[FileKind]map[protocol.CodeActionKind]bool{
			Go: {
				protocol.SourceOrganizeImports: true,
				protocol.QuickFix:              true,
			},
			Mod: {
				protocol.SourceOrganizeImports: true,
			},
			Sum: {},
		},
		SupportedCommands: []string{
			"tidy", // for go.mod files
		},
		Completion: CompletionOptions{
			Documentation: true,
			Deep:          true,
			FuzzyMatching: true,
			Literal:       true,
			Budget:        100 * time.Millisecond,
		},
		ComputeEdits: myers.ComputeEdits,
		Analyzers:    defaultAnalyzers,
		GoDiff:       true,
		LinkTarget:   "pkg.go.dev",
		TempModfile:  false,
	}
)

type Options struct {
	// Env is the current set of environment overrides on this view.
	Env []string

	// BuildFlags is used to adjust the build flags applied to the view.
	BuildFlags []string

	HoverKind        HoverKind
	DisabledAnalyses map[string]struct{}

	StaticCheck bool
	GoDiff      bool

	WatchFileChanges              bool
	InsertTextFormat              protocol.InsertTextFormat
	ConfigurationSupported        bool
	DynamicConfigurationSupported bool
	DynamicWatchedFilesSupported  bool
	PreferredContentFormat        protocol.MarkupKind
	LineFoldingOnly               bool

	SupportedCodeActions map[FileKind]map[protocol.CodeActionKind]bool

	SupportedCommands []string

	// TODO: Remove the option once we are certain there are no issues here.
	TextDocumentSyncKind protocol.TextDocumentSyncKind

	Completion CompletionOptions

	ComputeEdits diff.ComputeEdits

	Analyzers map[string]*analysis.Analyzer

	// LocalPrefix is used to specify goimports's -local behavior.
	LocalPrefix string

	VerboseOutput bool

	// WARNING: This configuration will be changed in the future.
	// It only exists while this feature is under development.
	// Disable use of the -modfile flag in Go 1.14.
	TempModfile bool

	LinkTarget string
}

type CompletionOptions struct {
	Deep              bool
	FuzzyMatching     bool
	CaseSensitive     bool
	Unimported        bool
	Documentation     bool
	FullDocumentation bool
	Placeholders      bool
	Literal           bool

	// Budget is the soft latency goal for completion requests. Most
	// requests finish in a couple milliseconds, but in some cases deep
	// completions can take much longer. As we use up our budget we
	// dynamically reduce the search scope to ensure we return timely
	// results. Zero means unlimited.
	Budget time.Duration
}

type HoverKind int

const (
	SingleLine = HoverKind(iota)
	NoDocumentation
	SynopsisDocumentation
	FullDocumentation

	// structured is an experimental setting that returns a structured hover format.
	// This format separates the signature from the documentation, so that the client
	// can do more manipulation of these fields.
	//
	// This should only be used by clients that support this behavior.
	Structured
)

type OptionResults []OptionResult

type OptionResult struct {
	Name  string
	Value interface{}
	Error error

	State       OptionState
	Replacement string
}

type OptionState int

const (
	OptionHandled = OptionState(iota)
	OptionDeprecated
	OptionUnexpected
)

type LinkTarget string

func SetOptions(options *Options, opts interface{}) OptionResults {
	var results OptionResults
	switch opts := opts.(type) {
	case nil:
	case map[string]interface{}:
		for name, value := range opts {
			results = append(results, options.set(name, value))
		}
	default:
		results = append(results, OptionResult{
			Value: opts,
			Error: errors.Errorf("Invalid options type %T", opts),
		})
	}
	return results
}

func (o *Options) ForClientCapabilities(caps protocol.ClientCapabilities) {
	// Check if the client supports snippets in completion items.
	if c := caps.TextDocument.Completion; c.CompletionItem.SnippetSupport {
		o.InsertTextFormat = protocol.SnippetTextFormat
	}
	// Check if the client supports configuration messages.
	o.ConfigurationSupported = caps.Workspace.Configuration
	o.DynamicConfigurationSupported = caps.Workspace.DidChangeConfiguration.DynamicRegistration
	o.DynamicWatchedFilesSupported = caps.Workspace.DidChangeWatchedFiles.DynamicRegistration

	// Check which types of content format are supported by this client.
	if hover := caps.TextDocument.Hover; len(hover.ContentFormat) > 0 {
		o.PreferredContentFormat = hover.ContentFormat[0]
	}
	// Check if the client supports only line folding.
	fr := caps.TextDocument.FoldingRange
	o.LineFoldingOnly = fr.LineFoldingOnly
}

func (o *Options) set(name string, value interface{}) OptionResult {
	result := OptionResult{Name: name, Value: value}
	switch name {
	case "env":
		menv, ok := value.(map[string]interface{})
		if !ok {
			result.errorf("invalid config gopls.env type %T", value)
			break
		}
		for k, v := range menv {
			o.Env = append(o.Env, fmt.Sprintf("%s=%s", k, v))
		}

	case "buildFlags":
		iflags, ok := value.([]interface{})
		if !ok {
			result.errorf("invalid config gopls.buildFlags type %T", value)
			break
		}
		flags := make([]string, 0, len(iflags))
		for _, flag := range iflags {
			flags = append(flags, fmt.Sprintf("%s", flag))
		}
		o.BuildFlags = flags

	case "noIncrementalSync":
		if v, ok := result.asBool(); ok && v {
			o.TextDocumentSyncKind = protocol.Full
		}
	case "watchFileChanges":
		result.setBool(&o.WatchFileChanges)
	case "completionDocumentation":
		result.setBool(&o.Completion.Documentation)
	case "usePlaceholders":
		result.setBool(&o.Completion.Placeholders)
	case "deepCompletion":
		result.setBool(&o.Completion.Deep)
	case "fuzzyMatching":
		result.setBool(&o.Completion.FuzzyMatching)
	case "caseSensitiveCompletion":
		result.setBool(&o.Completion.CaseSensitive)
	case "completeUnimported":
		result.setBool(&o.Completion.Unimported)
	case "completionBudget":
		if v, ok := result.asString(); ok {
			d, err := time.ParseDuration(v)
			if err != nil {
				result.errorf("failed to parse duration %q: %v", v, err)
				break
			}
			o.Completion.Budget = d
		}

	case "hoverKind":
		hoverKind, ok := value.(string)
		if !ok {
			result.errorf("invalid type %T for string option %q", value, name)
			break
		}
		switch hoverKind {
		case "NoDocumentation":
			o.HoverKind = NoDocumentation
		case "SingleLine":
			o.HoverKind = SingleLine
		case "SynopsisDocumentation":
			o.HoverKind = SynopsisDocumentation
		case "FullDocumentation":
			o.HoverKind = FullDocumentation
		case "Structured":
			o.HoverKind = Structured
		default:
			result.errorf("Unsupported hover kind", tag.Of("HoverKind", hoverKind))
		}

	case "linkTarget":
		linkTarget, ok := value.(string)
		if !ok {
			result.errorf("invalid type %T for string option %q", value, name)
			break
		}
		o.LinkTarget = linkTarget

	case "experimentalDisabledAnalyses":
		disabledAnalyses, ok := value.([]interface{})
		if !ok {
			result.errorf("Invalid type %T for []string option %q", value, name)
			break
		}
		o.DisabledAnalyses = make(map[string]struct{})
		for _, a := range disabledAnalyses {
			o.DisabledAnalyses[fmt.Sprint(a)] = struct{}{}
		}

	case "staticcheck":
		result.setBool(&o.StaticCheck)

	case "go-diff":
		result.setBool(&o.GoDiff)

	case "local":
		localPrefix, ok := value.(string)
		if !ok {
			result.errorf("invalid type %T for string option %q", value, name)
			break
		}
		o.LocalPrefix = localPrefix

	case "verboseOutput":
		result.setBool(&o.VerboseOutput)

	case "tempModfile":
		result.setBool(&o.TempModfile)

	// Deprecated settings.
	case "wantSuggestedFixes":
		result.State = OptionDeprecated

	case "disableDeepCompletion":
		result.State = OptionDeprecated
		result.Replacement = "deepCompletion"

	case "disableFuzzyMatching":
		result.State = OptionDeprecated
		result.Replacement = "fuzzyMatching"

	case "wantCompletionDocumentation":
		result.State = OptionDeprecated
		result.Replacement = "completionDocumentation"

	case "wantUnimportedCompletions":
		result.State = OptionDeprecated
		result.Replacement = "completeUnimported"

	default:
		result.State = OptionUnexpected
	}
	return result
}

func (r *OptionResult) errorf(msg string, values ...interface{}) {
	r.Error = errors.Errorf(msg, values...)
}

func (r *OptionResult) asBool() (bool, bool) {
	b, ok := r.Value.(bool)
	if !ok {
		r.errorf("Invalid type %T for bool option %q", r.Value, r.Name)
		return false, false
	}
	return b, true
}

func (r *OptionResult) asString() (string, bool) {
	b, ok := r.Value.(string)
	if !ok {
		r.errorf("Invalid type %T for string option %q", r.Value, r.Name)
		return "", false
	}
	return b, true
}

func (r *OptionResult) setBool(b *bool) {
	if v, ok := r.asBool(); ok {
		*b = v
	}
}

var defaultAnalyzers = map[string]*analysis.Analyzer{
	// The traditional vet suite:
	asmdecl.Analyzer.Name:      asmdecl.Analyzer,
	assign.Analyzer.Name:       assign.Analyzer,
	atomic.Analyzer.Name:       atomic.Analyzer,
	atomicalign.Analyzer.Name:  atomicalign.Analyzer,
	bools.Analyzer.Name:        bools.Analyzer,
	buildtag.Analyzer.Name:     buildtag.Analyzer,
	cgocall.Analyzer.Name:      cgocall.Analyzer,
	composite.Analyzer.Name:    composite.Analyzer,
	copylock.Analyzer.Name:     copylock.Analyzer,
	httpresponse.Analyzer.Name: httpresponse.Analyzer,
	loopclosure.Analyzer.Name:  loopclosure.Analyzer,
	lostcancel.Analyzer.Name:   lostcancel.Analyzer,
	nilfunc.Analyzer.Name:      nilfunc.Analyzer,
	printf.Analyzer.Name:       printf.Analyzer,
	shift.Analyzer.Name:        shift.Analyzer,
	stdmethods.Analyzer.Name:   stdmethods.Analyzer,
	structtag.Analyzer.Name:    structtag.Analyzer,
	tests.Analyzer.Name:        tests.Analyzer,
	unmarshal.Analyzer.Name:    unmarshal.Analyzer,
	unreachable.Analyzer.Name:  unreachable.Analyzer,
	unsafeptr.Analyzer.Name:    unsafeptr.Analyzer,
	unusedresult.Analyzer.Name: unusedresult.Analyzer,

	// Non-vet analyzers
	sortslice.Analyzer.Name: sortslice.Analyzer,
}
