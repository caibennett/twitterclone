PRAGMA foreign_keys = ON;

CREATE TABLE users (
    id integer PRIMARY KEY,
    username text NOT NULL,
    password text NOT NULL,

    name text,

    created_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_users_username
ON users (username);

CREATE TABLE posts (
    id integer PRIMARY KEY,
    user_id integer NOT NULL,

    content text NOT NULL,

    created_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at integer DEFAULT CURRENT_TIMESTAMP NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE sessions (
    token text PRIMARY KEY,
    user_id integer NOT NULL,
    ip_address text NOT NULL,
    expire_at integer NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX idx_sessions_token
ON sessions (token);
CREATE INDEX  idx_sessions_expire_at
ON sessions (expire_at);