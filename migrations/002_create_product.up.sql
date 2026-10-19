CREATE TABLE IF NOT EXISTS products (
    "id" BIGSERIAL PRIMARY KEY,
    "name" VARCHAR(250) NOT NULL,
    "subcategory_id" BIGINT NOT NULL,
    "description_dictionary_id" BIGINT,
    "price" real NOT NULL,
    "discount_price" real,
    "features_dictionary_id" BIGINT,
    "stock" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,
   
    CONSTRAINT description_dictionary_id_product_fk FOREIGN KEY ("description_dictionary_id") REFERENCES dictionary (id),
    CONSTRAINT subcategory_id_product_fk FOREIGN KEY ("subcategory_id") REFERENCES subcategories (id),
    CONSTRAINT features_dictionary_id_product_fk FOREIGN KEY ("features_dictionary_id") REFERENCES dictionary (id)
    );

CREATE TABLE IF NOT EXISTS image_products (
    "id" bigserial PRIMARY KEY,
    "product_id" bigint NOT NULL,
    "image_path" character varying(250) NOT NULL,
    "is_main" boolean DEFAULT FALSE,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
     
    CONSTRAINT product_id_image_products_fk FOREIGN KEY ("product_id") REFERENCES products("id")
    );
