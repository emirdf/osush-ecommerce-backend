CREATE TABLE device_tokens (
    "id" bigserial PRIMARY KEY,
    "client_id" bigint,
    "device_token" TEXT NOT NULL UNIQUE,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,

    CONSTRAINT client_id_device_tokens_fk FOREIGN KEY ("client_id") REFERENCES clients (id)
);

CREATE TABLE notifications (
    "id" bigserial PRIMARY KEY,
    "title_dictionary_id" BIGINT NOT NULL,
    "body_dictionary_id" BIGINT NOT NULL,
    "life" bigint NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,

    CONSTRAINT title_dictionary_id_notifications_fk FOREIGN KEY ("title_dictionary_id") REFERENCES dictionary (id),
    CONSTRAINT body_dictionary_id_notifications_fk FOREIGN KEY ("body_dictionary_id") REFERENCES dictionary (id)
);

CREATE TABLE client_notifications (
    "id" bigserial PRIMARY KEY,
    "device_token_id" bigint NOT NULL,
    "notification_id" bigint NOT NULL,
    "is_read" BOOLEAN DEFAULT FALSE,

    CONSTRAINT notification_id_client_notifications_fk FOREIGN KEY ("notification_id") REFERENCES notifications("id")
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT device_token_id_client_notifications_fk FOREIGN KEY ("device_token_id") REFERENCES device_tokens("id"),
    CONSTRAINT unique_client_notifications UNIQUE ("device_token_id", "notification_id")
);
