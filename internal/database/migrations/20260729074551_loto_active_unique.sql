-- Create index "lotogame_one_active_per_guild" to table: "loto_games"
CREATE UNIQUE INDEX "lotogame_one_active_per_guild" ON "loto_games" ("guild_id") WHERE active;
