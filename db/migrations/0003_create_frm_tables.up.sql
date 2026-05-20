CREATE TABLE frm_fields (
    id         BIGSERIAL PRIMARY KEY,
    field_id   INTEGER NULL,
    site_id    BIGINT NOT NULL,
    key        VARCHAR(255) NULL,
    type       VARCHAR(255) NULL,
    label      VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_fields_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE INDEX frm_fields_site_id_index ON frm_fields (site_id);

CREATE TABLE frm_entry_update_types (
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(255) NULL,
    title      VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE frm_entry_history (
    id             BIGSERIAL PRIMARY KEY,
    entry_id       INTEGER NULL,
    site_id        BIGINT NOT NULL,
    field_id       BIGINT NOT NULL,
    user_id        INTEGER NULL,
    update_type_id BIGINT NULL,
    old_value      TEXT NULL,
    new_value      TEXT NULL,
    change_date    TIMESTAMP NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_entry_history_site_id_fk        FOREIGN KEY (site_id)        REFERENCES sites (id)                  ON DELETE CASCADE,
    CONSTRAINT frm_entry_history_field_id_fk       FOREIGN KEY (field_id)       REFERENCES frm_fields (id)             ON DELETE CASCADE,
    CONSTRAINT frm_entry_history_update_type_id_fk FOREIGN KEY (update_type_id) REFERENCES frm_entry_update_types (id) ON DELETE SET NULL
);
CREATE INDEX frm_entry_history_site_id_index  ON frm_entry_history (site_id);
CREATE INDEX frm_entry_history_entry_id_index ON frm_entry_history (entry_id);
CREATE INDEX frm_entry_history_field_id_index ON frm_entry_history (field_id);
