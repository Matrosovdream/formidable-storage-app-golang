CREATE TABLE frm_easypost_shipment_history (
    id                   BIGSERIAL PRIMARY KEY,
    easypost_shipment_id VARCHAR(255) NULL,
    user_id              INTEGER NULL,
    site_id              BIGINT NOT NULL,
    change_type          VARCHAR(255) NULL,
    description          TEXT NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_shipment_history_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE INDEX frm_ep_shipment_history_site_id_index ON frm_easypost_shipment_history (site_id);
