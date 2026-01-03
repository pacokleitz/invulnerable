-- Rollback migration 005: Remove onlyFixed filters from webhook configurations

ALTER TABLE imagescan_webhook_configs
DROP COLUMN IF EXISTS status_change_only_fixed,
DROP COLUMN IF EXISTS scan_only_fixed;
