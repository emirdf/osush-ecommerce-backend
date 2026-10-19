SET client_encoding TO 'UTF-8';

CREATE TABLE IF NOT EXISTS dictionary (
    "id" bigserial PRIMARY KEY,
    "tm" varchar(250) NOT NULL,
    "ru" varchar(250) NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL
);

    CREATE TABLE  IF NOT EXISTS users (
    "id" bigserial primary key,
    "name" character varying(250) NOT NULL,
    "password" varchar(255) NOT NULL,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL
    );

INSERT INTO users (name, password)
    VALUES ('admin', '$2a$10$kHF6nJ0E33dVW.2iLY0nkugTrq.glAuACngt78Z6Gu2aXBJF0vloa');

    CREATE TABLE IF NOT EXISTS categories (
    "id" bigserial PRIMARY KEY,
    "name" varchar(255) NOT NULL,
    "image_path" character varying(250),
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL
);

CREATE TABLE  IF NOT EXISTS subcategories (
    "id" bigserial primary key,
    "category_id" bigint NOT NULL,
    "name" varchar(255) NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,

    CONSTRAINT category_id_subcategories_fk FOREIGN KEY ("category_id") REFERENCES categories("id")
    );

CREATE TABLE IF NOT EXISTS blacklist (
    "id" bigserial PRIMARY KEY,
    "token" varchar(250) NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
