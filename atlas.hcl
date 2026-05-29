env "local" {
  url = getenv("DATABASE_URL")
  dev = getenv("DATABASE_URL")
  migration {
    dir = "file://internal/database/migrations"
  }
}
