-- Migration 006: Rename only_fixed to only_fixable for clarity
-- "Fixable" is clearer than "fixed" - it means "has a fix available" not "remediated"

ALTER TABLE imagescan_webhook_configs
RENAME COLUMN scan_only_fixed TO scan_only_fixable;

ALTER TABLE imagescan_webhook_configs
RENAME COLUMN status_change_only_fixed TO status_change_only_fixable;

COMMENT ON COLUMN imagescan_webhook_configs.scan_only_fixable IS 'When true, only send scan completion notifications for vulnerabilities with fixes available';
COMMENT ON COLUMN imagescan_webhook_configs.status_change_only_fixable IS 'When true, only send status change notifications for vulnerabilities with fixes available';
