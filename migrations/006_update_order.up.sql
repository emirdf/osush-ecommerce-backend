DROP TABLE IF EXISTS order_products;
DROP TABLE IF EXISTS orders;


CREATE TABLE IF NOT EXISTS orders (
    "id" bigserial PRIMARY KEY,
    "client_id" bigint NOT NULL,
    "customer_name" varchar(250),
    "phone_number" varchar(250) NOT NULL,
    "city" varchar(250) NOT NULL,
    "delivery_address" varchar(250) NOT NULL,
    "notes" varchar(250),
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    CONSTRAINT client_id_order_fk FOREIGN KEY ("client_id") REFERENCES clients(id)
    );


CREATE TABLE IF NOT EXISTS order_products (
    "id" bigserial PRIMARY KEY,
    "order_id" bigint NOT NULL,
    "product_id" bigint NOT NULL,
    "quantity" int NOT NULL CHECK (quantity > 0),
    "price" real NOT NULL CHECK (price >= 0),
    "discount_price" real,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    CONSTRAINT order_id_order_products_fk FOREIGN KEY ("order_id") REFERENCES orders(id),

    CONSTRAINT product_id_order_products_fk FOREIGN KEY ("product_id") REFERENCES products(id),

    CONSTRAINT order_products_uq UNIQUE ("order_id", "product_id")
    );
