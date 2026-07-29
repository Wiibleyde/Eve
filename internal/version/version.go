// Package version exposes the build version of Eve.
//
// Version is injected at link time so that no source file has to be touched on
// release (the TS API hardcoded "0.1.0" and it rotted):
//
//	go build -ldflags "-X Eve/internal/version.Version=$(git describe --tags --always --dirty)" -o eve .
//
// Plain `go run .` / `go build .` builds keep the "dev" fallback.
package version

// Version is the build version reported by GET /api/v1/info.
//
// It must stay a plain package-level string variable with a constant
// initialiser: the linker's -X flag can only patch that shape.
var Version = "dev"
