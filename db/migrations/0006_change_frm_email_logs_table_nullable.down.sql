UPDATE frm_emails_log SET subject       = '' WHERE subject       IS NULL;
UPDATE frm_emails_log SET content_plain = '' WHERE content_plain IS NULL;
UPDATE frm_emails_log SET content_html  = '' WHERE content_html  IS NULL;
ALTER TABLE frm_emails_log ALTER COLUMN subject       SET NOT NULL;
ALTER TABLE frm_emails_log ALTER COLUMN content_plain SET NOT NULL;
ALTER TABLE frm_emails_log ALTER COLUMN content_html  SET NOT NULL;
