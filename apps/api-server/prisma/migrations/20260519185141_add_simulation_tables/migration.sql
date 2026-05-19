-- CreateTable
CREATE TABLE "sim_actions" (
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
    "timestamp" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "sim_leaderboards" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "user_id" TEXT NOT NULL,
    "username" TEXT NOT NULL,
    "metric" TEXT NOT NULL,
    "rank" INTEGER NOT NULL,
    "value" REAL NOT NULL,
    "updated_at" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "sim_anomalies" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "type" TEXT NOT NULL,
    "severity" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "data" TEXT NOT NULL,
    "detected_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_transactions" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "user_id" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "amount" INTEGER NOT NULL,
    "currency" TEXT NOT NULL,
    "status" TEXT NOT NULL DEFAULT 'pending',
    "external_ref" TEXT,
    "description" TEXT,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "transactions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);
INSERT INTO "new_transactions" ("amount", "created_at", "currency", "description", "id", "type", "user_id") SELECT "amount", "created_at", "currency", "description", "id", "type", "user_id" FROM "transactions";
DROP TABLE "transactions";
ALTER TABLE "new_transactions" RENAME TO "transactions";
CREATE TABLE "new_users" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "username" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "nickname" TEXT,
    "avatar" TEXT,
    "gold" INTEGER NOT NULL DEFAULT 0,
    "bb" INTEGER NOT NULL DEFAULT 0,
    "status" TEXT NOT NULL DEFAULT 'active',
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" DATETIME NOT NULL,
    "hands_played" INTEGER NOT NULL DEFAULT 0,
    "hands_won" INTEGER NOT NULL DEFAULT 0,
    "vpip" INTEGER NOT NULL DEFAULT 0,
    "pfr" INTEGER NOT NULL DEFAULT 0,
    "is_sim_user" BOOLEAN NOT NULL DEFAULT false,
    "sim_style" TEXT,
    "sim_personality" TEXT
);
INSERT INTO "new_users" ("avatar", "bb", "created_at", "email", "gold", "hands_played", "hands_won", "id", "nickname", "password", "pfr", "status", "updated_at", "username", "vpip") SELECT "avatar", "bb", "created_at", "email", "gold", "hands_played", "hands_won", "id", "nickname", "password", "pfr", "status", "updated_at", "username", "vpip" FROM "users";
DROP TABLE "users";
ALTER TABLE "new_users" RENAME TO "users";
CREATE UNIQUE INDEX "users_username_key" ON "users"("username");
CREATE UNIQUE INDEX "users_email_key" ON "users"("email");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;

-- CreateIndex
CREATE INDEX "sim_actions_user_id_timestamp_idx" ON "sim_actions"("user_id", "timestamp");

-- CreateIndex
CREATE INDEX "sim_actions_session_id_timestamp_idx" ON "sim_actions"("session_id", "timestamp");

-- CreateIndex
CREATE INDEX "sim_leaderboards_metric_rank_idx" ON "sim_leaderboards"("metric", "rank");

-- CreateIndex
CREATE UNIQUE INDEX "sim_leaderboards_metric_user_id_key" ON "sim_leaderboards"("metric", "user_id");

-- CreateIndex
CREATE INDEX "sim_anomalies_type_detected_at_idx" ON "sim_anomalies"("type", "detected_at");
