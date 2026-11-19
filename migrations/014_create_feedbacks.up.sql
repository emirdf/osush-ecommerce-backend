-- Feedback (suggestions) sent from the mobile app. Guests can send feedback
-- too, so client_id is optional. When a client deletes their account the
-- feedback stays and only loses the link to the client.
CREATE TABLE IF NOT EXISTS feedbacks (
    "id" bigserial PRIMARY KEY,
    "client_id" bigint,
    "message" TEXT NOT NULL,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,

    CONSTRAINT client_id_feedbacks_fk FOREIGN KEY ("client_id") REFERENCES clients (id) ON DELETE SET NULL,
    CONSTRAINT message_feedbacks_not_blank CHECK (length(btrim("message")) > 0)
);

CREATE INDEX IF NOT EXISTS feedbacks_created_at_idx ON feedbacks ("created_at" DESC);
CREATE INDEX IF NOT EXISTS feedbacks_client_id_idx ON feedbacks ("client_id");
