package main

import (
	"github.com/Smana/scia/cmd"
)

// Version information injected by GoReleaser during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Set version information for cmd package
	cmd.SetVersionInfo(version, commit, date, builtBy)
	cmd.Execute()
}
