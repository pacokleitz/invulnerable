-- Rollback migration 006: Rename columns back to only_fixed

ALTER TABLE imagescan_webhook_configs
RENAME COLUMN status_change_only_fixable TO status_change_only_fixed;

ALTER TABLE imagescan_webhook_configs
RENAME COLUMN scan_only_fixable TO scan_only_fixed;

COMMENT ON COLUMN imagescan_webhook_configs.scan_only_fixed IS 'When true, only send scan completion notifications for vulnerabilities with fixes available';
COMMENT ON COLUMN imagescan_webhook_configs.status_change_only_fixed IS 'When true, only send status change notifications for vulnerabilities with fixes available';
