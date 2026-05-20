CREATE INDEX idx_ep_hist_site_shipment_created ON frm_easypost_shipment_history (site_id, easypost_shipment_id, created_at);
CREATE INDEX idx_ep_hist_site_created           ON frm_easypost_shipment_history (site_id, created_at);
CREATE INDEX idx_ep_hist_site_change_type       ON frm_easypost_shipment_history (site_id, change_type);
CREATE INDEX idx_ep_hist_site_user              ON frm_easypost_shipment_history (site_id, user_id);
