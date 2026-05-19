-- CreateTable
CREATE TABLE "users" (
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
    "pfr" INTEGER NOT NULL DEFAULT 0
);

-- CreateTable
CREATE TABLE "refresh_tokens" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "token" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "expires_at" DATETIME NOT NULL,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "refresh_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "clubs" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "owner_id" TEXT NOT NULL,
    "join_type" TEXT NOT NULL DEFAULT 'approval',
    "status" TEXT NOT NULL DEFAULT 'active',
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "clubs_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "club_members" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "club_id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "role" TEXT NOT NULL DEFAULT 'member',
    "joined_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "club_members_club_id_fkey" FOREIGN KEY ("club_id") REFERENCES "clubs" ("id") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "club_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "club_tables" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "club_id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "stakes" TEXT NOT NULL DEFAULT '1/2',
    "game_type" TEXT NOT NULL DEFAULT 'cash',
    "status" TEXT NOT NULL DEFAULT 'active',
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "club_tables_club_id_fkey" FOREIGN KEY ("club_id") REFERENCES "clubs" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "hand_histories" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "user_id" TEXT NOT NULL,
    "table_id" TEXT NOT NULL,
    "hand_data" TEXT NOT NULL,
    "result" TEXT NOT NULL,
    "won_amount" INTEGER NOT NULL DEFAULT 0,
    "played_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "hand_histories_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "transactions" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "user_id" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    "amount" INTEGER NOT NULL,
    "currency" TEXT NOT NULL,
    "description" TEXT,
    "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "transactions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateIndex
CREATE UNIQUE INDEX "users_username_key" ON "users"("username");

-- CreateIndex
CREATE UNIQUE INDEX "users_email_key" ON "users"("email");

-- CreateIndex
CREATE UNIQUE INDEX "refresh_tokens_token_key" ON "refresh_tokens"("token");

-- CreateIndex
CREATE UNIQUE INDEX "club_members_club_id_user_id_key" ON "club_members"("club_id", "user_id");
