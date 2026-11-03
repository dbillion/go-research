package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/deeone/go-research/cmd/outreach"
)

// Version info set at build time via -ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	// Try to get version from build info if not set via ldflags
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					commit = setting.Value
				case "vcs.time":
					date = setting.Value
				}
			}
		}
	}
}

func main() {
	// Add version command
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("outreach version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	}

	if err := outreach.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}