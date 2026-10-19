DROP TABLE IF EXISTS order_products;
DROP TABLE IF EXISTS orders;


CREATE TABLE IF NOT EXISTS orders (
    "id" bigserial PRIMARY KEY,
    "client_id" bigint NOT NULL,
    "total_price" real NOT NULL,
    "address" varchar(250) NOT NULL,
    "order_time" TIMESTAMP NOT NULL,
    "status" varchar(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','APPROVED','CANCELED_BY_ADMIN','CANCELED_BY_CLIENT')),
    "reason" varchar(250),
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    CONSTRAINT client_id_order_fk FOREIGN KEY ("client_id") REFERENCES clients(id)
);

CREATE TABLE IF NOT EXISTS order_products (
    "id" bigserial PRIMARY KEY,
    "order_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "total_product" int NOT NULL CHECK (total_product > 0),
    "price" real NOT NULL CHECK (price >= 0),
    "discount_price" real,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    CONSTRAINT order_id_order_products_fk FOREIGN KEY ("order_id") REFERENCES orders(id),
    CONSTRAINT product_id_order_products_fk FOREIGN KEY ("product_id") REFERENCES products(id),
    CONSTRAINT order_products_uq UNIQUE ("order_id", "product_id")
);
