CREATE TABLE inventory
(
    art_id   VARCHAR(255) not null,
    art_name VARCHAR(255) NOT NULL,
    stock    INT          not null CHECK (stock >= 0),
    PRIMARY KEY (art_id)
);

CREATE TABLE product
(
    product_name VARCHAR(255) NOT NULL,
    art_id       VARCHAR(255) not null REFERENCES inventory (art_id),
    amount       INT          not null CHECK (amount > 0),
    PRIMARY KEY (product_name, art_id)
);
