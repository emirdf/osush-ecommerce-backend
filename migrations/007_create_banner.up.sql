CREATE TABLE IF NOT EXISTS banners (
    "id" BIGSERIAL PRIMARY KEY,
    "title_dictionary_id" BIGINT NOT NULL,
    "url" VARCHAR(250),
    "description_dictionary_id" BIGINT,
    "phone_number" VARCHAR(250),
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,
   
    CONSTRAINT description_dictionary_id_banners_fk FOREIGN KEY ("description_dictionary_id") REFERENCES dictionary (id),
    CONSTRAINT title_dictionary_id_banners_fk FOREIGN KEY ("title_dictionary_id") REFERENCES dictionary (id)
    );

CREATE TABLE IF NOT EXISTS image_banners (
    "id" bigserial PRIMARY KEY,
    "banner_id" bigint NOT NULL,
    "image_path" character varying(250) NOT NULL,
    "is_main" boolean DEFAULT FALSE,
    "created_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')  NOT NULL,
     
    CONSTRAINT banner_id_image_banners_fk FOREIGN KEY ("banner_id") REFERENCES banners("id")
    );
