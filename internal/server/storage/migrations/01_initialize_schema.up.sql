BEGIN;

CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY NOT NULL,
    login VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    modified_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS texts(
    id UUID PRIMARY KEY NOT NULL,
    text TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS binaries(
    id UUID PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    content_type TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS bank_cards(
    id UUID PRIMARY KEY NOT NULL,
    card_num TEXT NOT NULL,
    expiry TEXT NOT NULL,
    cvv INT NOT NULL
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