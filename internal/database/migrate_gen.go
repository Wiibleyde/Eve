//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	"ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	entschema "entgo.io/ent/dialect/sql/schema"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	entmigrate "Eve/internal/database/ent/migrate"
)

func main() {
	_ = godotenv.Load(".env")
	name := "changes"
	if len(os.Args) >= 2 {
		name = os.Args[1]
	}
	ctx := context.Background()
	dir, err := migrate.NewLocalDir("./internal/database/migrations")
	if err != nil {
		log.Fatal(err)
	}
	devURL := os.Getenv("DATABASE_URL")
	if devURL == "" {
		devURL = "postgres://postgres:postgres@localhost:5432/eve?sslmode=disable"
	}
	opts := []entschema.MigrateOption{
		entschema.WithDir(dir),
		entschema.WithMigrationMode(entschema.ModeInspect),
		entschema.WithDialect(dialect.Postgres),
		entschema.WithFormatter(migrate.DefaultFormatter),
	}
	if err := entmigrate.NamedDiff(ctx, devURL, name, opts...); err != nil {
		log.Fatal(err)
	}
}
