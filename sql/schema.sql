PRAGMA foreign_kes = ON;

CREATE TABLE IF NOT EXISTS users (
    id integer PRIMARY KEY,
    username text PRIMARY KEY,
    password text NOT NULL,

    created_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL
)

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username
ON users (username);