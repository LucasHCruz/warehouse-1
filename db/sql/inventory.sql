CREATE TABLE inventory(
   art_id VARCHAR(255) not null ,
   art_name VARCHAR(255) NOT NULL,
   stock  INT not null CHECK (stock >= 0),
   PRIMARY KEY(art_id)
);

CREATE TABLE product(
   product_name VARCHAR(255) NOT NULL,
   art_id VARCHAR(255) not null REFERENCES inventory (art_id),
   amount  INT not null CHECK (amount > 0),
   PRIMARY KEY(product_name,art_id)
);

INSERT INTO inventory(art_id, art_name, stock)
VALUES ('1','leg',12);

INSERT INTO product(product_name, art_id, amount)
VALUES ('Dining Chair','1',4);

SELECT pr.product_name, min(i.stock/pr.amount) as available_product --,i.art_id,pr.amount,i.stock
-- ,count(pr.product_name)
--Min(CASE WHEN i.stock > 0 THEN stock ELSE 0 END) AS Location_001
--   ,SUM(CASE WHEN i.stock > 0 THEN stock ELSE 0 END) AS Location_001
FROM product pr,inventory i
where pr.art_id=i.art_id AND i.stock>pr.amount
GROUP BY pr.product_name--, i.art_id,pr.amount,i.stock
order by pr.product_name;


UPDATE inventory i
SET stock=stock-1
from product pr
WHERE pr.art_id= i.art_id and
pr.product_name='Dining Chair';