-- Fails while orders of deleted clients (client_id IS NULL) exist; delete or
-- reassign those orders first.
ALTER TABLE orders
    DROP CONSTRAINT client_id_order_fk,
    ADD CONSTRAINT client_id_order_fk FOREIGN KEY ("client_id") REFERENCES clients(id);

ALTER TABLE orders ALTER COLUMN client_id SET NOT NULL;
