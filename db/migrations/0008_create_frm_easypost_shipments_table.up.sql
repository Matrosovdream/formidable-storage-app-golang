-- frm_easypost_shipments
CREATE TABLE frm_easypost_shipments (
    id                   BIGSERIAL PRIMARY KEY,
    easypost_shipment_id VARCHAR(255) NULL,
    entry_id             BIGINT NULL,
    site_id              BIGINT NOT NULL,
    is_return            BOOLEAN NOT NULL DEFAULT FALSE,
    status               VARCHAR(255) NULL,
    tracking_code        VARCHAR(255) NULL,
    tracking_url         VARCHAR(255) NULL,
    refund_status        VARCHAR(255) NULL,
    mode                 VARCHAR(20) NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_shipments_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uq_ep_shipments_site_shipment ON frm_easypost_shipments (site_id, easypost_shipment_id);
CREATE INDEX idx_ep_shipments_site_entry  ON frm_easypost_shipments (site_id, entry_id);
CREATE INDEX idx_ep_shipments_site_status ON frm_easypost_shipments (site_id, status);

-- frm_easypost_shipment_addresses
CREATE TABLE frm_easypost_shipment_addresses (
    id                   BIGSERIAL PRIMARY KEY,
    easypost_id          VARCHAR(255) NOT NULL,
    easypost_shipment_id VARCHAR(255) NOT NULL,
    entry_id             BIGINT NULL,
    site_id              BIGINT NOT NULL,
    address_type         VARCHAR(20) NOT NULL,
    name                 VARCHAR(255) NULL,
    company              VARCHAR(255) NULL,
    street1              VARCHAR(255) NULL,
    street2              VARCHAR(255) NULL,
    city                 VARCHAR(255) NULL,
    state                VARCHAR(50) NULL,
    zip                  VARCHAR(20) NULL,
    country              CHAR(2) NULL,
    phone                VARCHAR(255) NULL,
    email                VARCHAR(255) NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_addr_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uq_ep_addr_site_epid        ON frm_easypost_shipment_addresses (site_id, easypost_id);
CREATE INDEX        idx_ep_addr_site_shipment   ON frm_easypost_shipment_addresses (site_id, easypost_shipment_id);
CREATE INDEX        idx_ep_addr_site_entry      ON frm_easypost_shipment_addresses (site_id, entry_id);
CREATE INDEX        idx_ep_addr_site_ship_type  ON frm_easypost_shipment_addresses (site_id, easypost_shipment_id, address_type);

-- frm_easypost_shipment_labels
CREATE TABLE frm_easypost_shipment_labels (
    id                   BIGSERIAL PRIMARY KEY,
    easypost_id          VARCHAR(255) NOT NULL,
    easypost_shipment_id VARCHAR(255) NOT NULL,
    entry_id             BIGINT NULL,
    site_id              BIGINT NOT NULL,
    date_advance         INTEGER NOT NULL DEFAULT 0,
    integrated_form      VARCHAR(255) NULL,
    label_date           TIMESTAMP NULL,
    label_resolution     INTEGER NULL,
    label_size           VARCHAR(20) NULL,
    label_type           VARCHAR(50) NULL,
    label_file_type      VARCHAR(50) NULL,
    label_url            TEXT NULL,
    label_pdf_url        TEXT NULL,
    label_zpl_url        TEXT NULL,
    label_epl2_url       TEXT NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_label_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uq_ep_label_site_epid      ON frm_easypost_shipment_labels (site_id, easypost_id);
CREATE INDEX        idx_ep_label_site_shipment ON frm_easypost_shipment_labels (site_id, easypost_shipment_id);
CREATE INDEX        idx_ep_label_site_entry    ON frm_easypost_shipment_labels (site_id, entry_id);

-- frm_easypost_shipment_parcels
CREATE TABLE frm_easypost_shipment_parcels (
    id                   BIGSERIAL PRIMARY KEY,
    easypost_id          VARCHAR(255) NOT NULL,
    easypost_shipment_id VARCHAR(255) NOT NULL,
    entry_id             BIGINT NULL,
    site_id              BIGINT NOT NULL,
    length               NUMERIC(10,2) NULL,
    width                NUMERIC(10,2) NULL,
    height               NUMERIC(10,2) NULL,
    weight               NUMERIC(10,2) NULL,
    created_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_parcel_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uq_ep_parcel_site_epid      ON frm_easypost_shipment_parcels (site_id, easypost_id);
CREATE INDEX        idx_ep_parcel_site_shipment ON frm_easypost_shipment_parcels (site_id, easypost_shipment_id);
CREATE INDEX        idx_ep_parcel_site_entry    ON frm_easypost_shipment_parcels (site_id, entry_id);

-- frm_easypost_shipment_rates
CREATE TABLE frm_easypost_shipment_rates (
    id                       BIGSERIAL PRIMARY KEY,
    easypost_id              VARCHAR(255) NOT NULL,
    easypost_shipment_id     VARCHAR(255) NOT NULL,
    entry_id                 BIGINT NULL,
    site_id                  BIGINT NOT NULL,
    mode                     VARCHAR(20) NOT NULL DEFAULT 'test',
    service                  VARCHAR(100) NULL,
    carrier                  VARCHAR(100) NULL,
    rate                     NUMERIC(12,2) NULL,
    currency                 CHAR(3) NULL,
    retail_rate              NUMERIC(12,2) NULL,
    retail_currency          CHAR(3) NULL,
    list_rate                NUMERIC(12,2) NULL,
    list_currency            CHAR(3) NULL,
    billing_type             VARCHAR(50) NULL,
    delivery_days            INTEGER NULL,
    delivery_date            TIMESTAMP NULL,
    delivery_date_guaranteed BOOLEAN NULL,
    est_delivery_days        INTEGER NULL,
    created_at               TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT frm_ep_rate_site_id_fk FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX uq_ep_rate_site_epid              ON frm_easypost_shipment_rates (site_id, easypost_id);
CREATE INDEX        idx_ep_rate_site_shipment        ON frm_easypost_shipment_rates (site_id, easypost_shipment_id);
CREATE INDEX        idx_ep_rate_site_entry           ON frm_easypost_shipment_rates (site_id, entry_id);
CREATE INDEX        idx_ep_rate_site_carrier_service ON frm_easypost_shipment_rates (site_id, carrier, service);
CREATE INDEX        idx_ep_rate_site_ship_rate       ON frm_easypost_shipment_rates (site_id, easypost_shipment_id, rate);
