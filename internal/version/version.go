// Package version reports the build metadata of the catl binary: version,
// commits, and build date. Values are injected at build time via -ldflags.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Injected at build time via -ldflags.
var (
	version     = "dev"
	shortCommit = "none"
	fullCommit  = "none"
	date        = "unknown"
	dirty       = false
)

func init() {
	// Fall back to runtime/debug info when built via 'go install' or local 'go build'
	if info, ok := debug.ReadBuildInfo(); ok {
		if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if fullCommit == "none" {
					fullCommit = s.Value
				}
				if shortCommit == "none" {
					shortCommit = deriveShortCommit(s.Value)
				}
			case "vcs.time":
				if date == "unknown" {
					date = s.Value
				}
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
	}

	// If fullCommit was injected but shortCommit was not, derive it
	if shortCommit == "none" && fullCommit != "none" {
		shortCommit = deriveShortCommit(fullCommit)
	}
}

// deriveShortCommit derives a short commit hash from a full commit hash.
func deriveShortCommit(fullCommit string) string {
	var s string
	if len(fullCommit) >= 9 {
		s = fullCommit[:9]
	} else if len(fullCommit) >= 7 {
		s = fullCommit[:7]
	} else {
		s = fullCommit
	}
	return s
}

// Version returns the semantic version or pre-release tag (e.g. "0.1.0-alpha.1").
func Version() string {
	return version
}

// ShortCommit returns the abbreviated Git commit hash (7-9 characters) for summary displays.
func ShortCommit() string {
	return shortCommit
}

// FullCommit returns the complete 40-character Git commit hash for diagnostics and releases.
func FullCommit() string {
	return fullCommit
}

// Date returns the build timestamp.
func Date() string {
	return date
}

// String returns the canonical single-line summary: "catl <version> (<short-commit> <date>)".
func String() string {
	if shortCommit != "none" && date != "unknown" {
		d := date
		if len(d) > 10 {
			d = d[:10] // Extract YYYY-MM-DD from RFC3339
		}
		return fmt.Sprintf("catl %s (%s %s)", version, shortCommit, d)
	}
	return "catl " + version
}

// Detailed returns the Cargo-style key-value diagnostic output for verbose inspection.
func Detailed() string {
	var sb strings.Builder
	sb.WriteString(String())
	sb.WriteString("\nrelease: ")
	sb.WriteString(version)
	sb.WriteString("\ncommit-hash: ")
	sb.WriteString(fullCommit)
	sb.WriteString("\ncommit-date: ")
	sb.WriteString(date)
	fmt.Fprintf(&sb, "\nhost: %s/%s", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&sb, "\ncompiler: %s (%s)", runtime.Compiler, runtime.Version())
	if dirty {
		sb.WriteString("\ndirty: true")
	}
	return sb.String()
}
