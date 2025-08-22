package version

import (
	runtimeDebug "runtime/debug"
)

// When building with CI or Make, Version and Commit is set using `ldflags`
var (
	Version string
	Commit  string
)

func init() {
	// Only use buildInfo if version wasn't set by ldflags, that is its being build by `go install`
	if Version == "" {
		// Main.Version is based on the version control system tag or commit.
		// This useful when app is build with `go install`
		// See: https://antonz.org/go-1-24/#main-modules-version
		buildInfo, _ := runtimeDebug.ReadBuildInfo()
		Version = buildInfo.Main.Version
	}
}
