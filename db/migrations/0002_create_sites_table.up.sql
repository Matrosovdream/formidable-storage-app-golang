CREATE TABLE sites (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    url        VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT sites_url_unique UNIQUE (url)
);

CREATE TABLE site_tokens (
    id          BIGSERIAL PRIMARY KEY,
    site_id     BIGINT NOT NULL,
    token       VARCHAR(255) NOT NULL,
    valid_until TIMESTAMP NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT site_tokens_token_unique UNIQUE (token),
    CONSTRAINT site_tokens_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE INDEX site_tokens_site_id_index ON site_tokens (site_id);
