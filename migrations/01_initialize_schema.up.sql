BEGIN;

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY NOT NULL,
    login VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS items_data(
    id UUID PRIMARY KEY NOT NULL,
    data TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS metas(
    id UUID PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL,
    data_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL
);

CREATE INDEX login_ix ON users (login);
CREATE INDEX user_id_ix ON metas (user_id);
CREATE INDEX data_id_ix ON metas (data_id);

COMMIT ;