//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	err := entc.Generate("./tables", &gen.Config{
		Target:  "./ent",
		Package: "Eve/internal/database/ent",
		Features: []gen.Feature{
			gen.FeatureVersionedMigration,
		},
	})
	if err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
