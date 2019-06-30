package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.gophers.dev/cmds/petrify/v5"
)

func main() {
	cfg := parseArgs()
	err := petrify.Translate(cfg)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "petrify: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs create s a new, filled configuration instance
// by reading and parsing command line options.
//
// This function exits the program with an error, if
// any of the command line options are incorrect.
func parseArgs() *petrify.Config {
	var version bool

	c := petrify.NewConfig()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&c.Debug, "debug", c.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.BoolVar(&c.Dev, "dev", c.Dev, "Similar to debug, but does not emit absolute paths. Expects a rootDir variable to already exist in the generated code's package.")
	flag.StringVar(&c.Tags, "tags", c.Tags, "Optional set of build tags to include.")
	flag.StringVar(&c.Prefix, "prefix", c.Prefix, "Optional path prefix to strip off asset names.")
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in the generated code.")
	flag.BoolVar(&c.NoMemCopy, "nomemcopy", c.NoMemCopy, "Use a .rodata hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
	flag.BoolVar(&c.NoCompress, "nocompress", c.NoCompress, "Assets will *not* be GZIP compressed when this flag is specified.")
	flag.BoolVar(&c.NoMetadata, "nometadata", c.NoMetadata, "Assets will not preserve size, mode, and modtime info.")
	flag.UintVar(&c.Mode, "mode", c.Mode, "Optional file mode override for all files.")
	flag.Int64Var(&c.ModTime, "modtime", c.ModTime, "Optional modification unix timestamp override for all files.")
	flag.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")
	flag.BoolVar(&version, "version", false, "Displays version information.")

	var ignore []string
	flag.Var((*AppendSliceValue)(&ignore), "ignore", "Regex pattern to ignore")

	flag.Parse()

	var patterns []*regexp.Regexp
	for _, pattern := range ignore {
		patterns = append(patterns, regexp.MustCompile(pattern))
	}
	c.Ignore = patterns

	if version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	// Make sure we have input paths.
	if flag.NArg() == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Missing <input dir>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create input configurations.
	c.Input = make([]petrify.InputConfig, flag.NArg())
	for i := range c.Input {
		c.Input[i] = parseInput(flag.Arg(i))
	}

	return c
}

// parseRecursive determines whether the given path has a recrusive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  e.g.:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(path string) petrify.InputConfig {
	if strings.HasSuffix(path, "/...") {
		return petrify.InputConfig{
			Path:      filepath.Clean(path[:len(path)-4]),
			Recursive: true,
		}
	}
	return petrify.InputConfig{
		Path:      filepath.Clean(path),
		Recursive: false,
	}
}
