-- A client can delete their account. Their orders stay for the shop's records
-- (each order keeps its own customer name, phone and address) and lose only
-- the link to the deleted client.
ALTER TABLE orders ALTER COLUMN client_id DROP NOT NULL;

ALTER TABLE orders
    DROP CONSTRAINT client_id_order_fk,
    ADD CONSTRAINT client_id_order_fk FOREIGN KEY ("client_id") REFERENCES clients(id) ON DELETE SET NULL;
