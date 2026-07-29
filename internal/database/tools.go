//go:build tools

package database

// entc.go carries `//go:build ignore`, so `go mod tidy` cannot see the codegen
// dependencies it needs and drops them from go.mod/go.sum, which breaks
// `go generate ./internal/database/...`. These blank imports keep them pinned.
import (
	_ "entgo.io/ent/entc"
	_ "entgo.io/ent/entc/gen"
)
