package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rafaelespinoza/alf"
)

// Pieces of version metadata that can be set through -ldflags at build time.
var (
	versionBranchName string
	versionBuildTime  string
	versionCommitHash string
	versionGoVersion  string
	versionTag        string
)

func makeVersion(name string) alf.Directive {
	var outputFormat string
	return &alf.Command{
		Description: "show metadata about the build",
		Setup: func(p flag.FlagSet) *flag.FlagSet {
			flags := newFlagSet(name)
			flags.StringVar(&outputFormat, "format", "tsv", `output format (json|tsv)`)
			flags.Usage = func() {
				fmt.Fprintf(flags.Output(), `Print some versioning info about the build.
`)
				printFlagDefaults(flags)
			}
			return flags
		},
		Run: func(ctx context.Context) error {
			return runVersion(outputFormat, os.Stdout)
		},
	}
}

func runVersion(outputFormat string, w io.Writer) error {
	versionData := []struct{ Key, Val string }{
		{"BranchName", versionBranchName},
		{"BuildTime", versionBuildTime},
		{"CommitHash", versionCommitHash},
		{"GoVersion", versionGoVersion},
		{"Tag", versionTag},
	}

	switch strings.ToLower(strings.TrimSpace(outputFormat)) {
	case "json":
		versionDataMap := make(map[string]string, len(versionData))
		for _, tuple := range versionData {
			versionDataMap[tuple.Key] = tuple.Val
		}

		out, err := json.Marshal(versionDataMap)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(w, "%s\n", out)
		return nil
	default:
		tw := tabwriter.NewWriter(w, 8, 4, 1, '\t', 0)
		for _, tuple := range versionData {
			_, _ = fmt.Fprintf(tw, "%s:\t%s\n", tuple.Key, tuple.Val)
		}
		return tw.Flush()
	}
}
