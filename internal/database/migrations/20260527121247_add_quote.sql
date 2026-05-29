-- Create "quotes" table
CREATE TABLE "quotes" ("id" character varying NOT NULL, "guild_id" character varying NOT NULL, "author_id" character varying NOT NULL, "quote" character varying NOT NULL, "context" character varying NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NOT NULL, PRIMARY KEY ("id"));
