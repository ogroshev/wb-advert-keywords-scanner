CREATE TABLE advert_keyword
(
    id                BIGSERIAL PRIMARY KEY,
    advert_company_id BIGINT NOT NULL,
    keyword           TEXT   NOT NULL,
    category        VARCHAR(255) NOT NULL,
    create_dt         TIMESTAMP DEFAULT NOW(),
    update_dt         TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY (advert_company_id) REFERENCES advert_company (id),
    UNIQUE (advert_company_id, keyword)
);
