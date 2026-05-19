-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_sim_actions" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "session_id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "table_id" TEXT NOT NULL,
    "hand_number" INTEGER NOT NULL,
    "phase" TEXT NOT NULL,
    "action" TEXT NOT NULL,
    "amount" INTEGER NOT NULL DEFAULT 0,
    "pot_before" INTEGER NOT NULL,
    "pot_after" INTEGER NOT NULL,
    "stack_before" INTEGER NOT NULL,
    "stack_after" INTEGER NOT NULL,
    "hole_cards" TEXT,
    "community" TEXT,
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "sim_actions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_sim_actions" ("action", "amount", "community", "hand_number", "hole_cards", "id", "phase", "pot_after", "pot_before", "session_id", "stack_after", "stack_before", "table_id", "timestamp", "user_id") SELECT "action", "amount", "community", "hand_number", "hole_cards", "id", "phase", "pot_after", "pot_before", "session_id", "stack_after", "stack_before", "table_id", "timestamp", "user_id" FROM "sim_actions";
DROP TABLE "sim_actions";
ALTER TABLE "new_sim_actions" RENAME TO "sim_actions";
CREATE INDEX "sim_actions_user_id_timestamp_idx" ON "sim_actions"("user_id", "timestamp");
CREATE INDEX "sim_actions_session_id_timestamp_idx" ON "sim_actions"("session_id", "timestamp");
CREATE TABLE "new_sim_leaderboards" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "user_id" TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "metric" TEXT NOT NULL,
    "rank" INTEGER NOT NULL,
    "value" REAL NOT NULL,
    "updated_at" DATETIME NOT NULL,
    CONSTRAINT "sim_leaderboards_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_sim_leaderboards" ("id", "metric", "rank", "updated_at", "user_id", "username", "value") SELECT "id", "metric", "rank", "updated_at", "user_id", "username", "value" FROM "sim_leaderboards";
DROP TABLE "sim_leaderboards";
ALTER TABLE "new_sim_leaderboards" RENAME TO "sim_leaderboards";
CREATE INDEX "sim_leaderboards_metric_rank_idx" ON "sim_leaderboards"("metric", "rank");
CREATE UNIQUE INDEX "sim_leaderboards_metric_user_id_key" ON "sim_leaderboards"("metric", "user_id");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
