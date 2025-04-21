CREATE TABLE users
(
    id            UUID PRIMARY KEY,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL
);
CREATE TABLE pvz
(
    id                UUID PRIMARY KEY,
    city              TEXT        NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    registration_date TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE reception
(
    id        UUID PRIMARY KEY,
    pvz_id    UUID REFERENCES pvz (id),
    date_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    status    TEXT        NOT NULL CHECK (status IN ('in_progress', 'close'))
);
CREATE UNIQUE INDEX one_open_reception ON reception (pvz_id) WHERE status='in_progress';
CREATE TABLE product
(
    id           UUID PRIMARY KEY,
    reception_id UUID REFERENCES reception (id),
    date_time    TIMESTAMPTZ NOT NULL DEFAULT now(),
    type         TEXT        NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь'))
);