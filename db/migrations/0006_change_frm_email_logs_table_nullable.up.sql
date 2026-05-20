ALTER TABLE frm_emails_log ALTER COLUMN subject       DROP NOT NULL;
ALTER TABLE frm_emails_log ALTER COLUMN content_plain DROP NOT NULL;
ALTER TABLE frm_emails_log ALTER COLUMN content_html  DROP NOT NULL;
