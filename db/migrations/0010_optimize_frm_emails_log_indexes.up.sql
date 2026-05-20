CREATE INDEX idx_eml_site_entry        ON frm_emails_log (site_id, entry_id);
CREATE INDEX idx_eml_site_form         ON frm_emails_log (site_id, form_id);
CREATE INDEX idx_eml_site_date_sent    ON frm_emails_log (site_id, date_sent);
CREATE INDEX idx_eml_site_status_date  ON frm_emails_log (site_id, status, date_sent);
CREATE INDEX idx_eml_site_mailer_date  ON frm_emails_log (site_id, mailer, date_sent);
CREATE INDEX idx_eml_site_original     ON frm_emails_log (site_id, original_log_id);
CREATE INDEX idx_eml_message_id        ON frm_emails_log (message_id);
CREATE INDEX idx_eml_site_message_id   ON frm_emails_log (site_id, message_id);
