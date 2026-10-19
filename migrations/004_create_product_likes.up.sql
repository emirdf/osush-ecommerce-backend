CREATE TABLE product_likes (
    "client_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,

    CONSTRAINT client_id_product_likes_fk FOREIGN KEY ("client_id") REFERENCES clients("id"),
    CONSTRAINT product_id_product_likes_fk FOREIGN KEY ("product_id") REFERENCES products("id"),
    CONSTRAINT product_likes_client_product_uq UNIQUE ("client_id", "product_id")
);
