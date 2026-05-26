-- Partial unique index on (site_id, field_id) for ON CONFLICT upserts.
-- field_id is nullable in the schema, so the constraint only applies to rows
-- that have a value. Generator and seed paths always set field_id.
CREATE UNIQUE INDEX IF NOT EXISTS frm_fields_site_field_uidx
    ON frm_fields (site_id, field_id)
    WHERE field_id IS NOT NULL;
