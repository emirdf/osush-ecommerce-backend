ALTER TABLE image_products
DROP CONSTRAINT product_id_image_products_fk;

ALTER TABLE image_products
ADD CONSTRAINT product_id_image_products_fk
FOREIGN KEY ("product_id")
REFERENCES products("id");


ALTER TABLE basket
DROP CONSTRAINT product_id_basket_fk;

ALTER TABLE basket
ADD CONSTRAINT product_id_basket_fk
FOREIGN KEY ("product_id")
REFERENCES products("id");


ALTER TABLE order_products
DROP CONSTRAINT product_id_order_products_fk;

ALTER TABLE order_products
ADD CONSTRAINT product_id_order_products_fk
FOREIGN KEY ("product_id")
REFERENCES products("id");


ALTER TABLE product_likes
DROP CONSTRAINT product_id_product_likes_fk;

ALTER TABLE product_likes
ADD CONSTRAINT product_id_product_likes_fk
FOREIGN KEY ("product_id")
REFERENCES products("id");
