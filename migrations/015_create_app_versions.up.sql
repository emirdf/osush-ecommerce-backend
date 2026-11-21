-- Mobile app versions per platform, edited from the admin panel.
--   min_version    : older apps must update (force update, blocking dialog)
--   latest_version : newest release; older apps are told about it
--                    (optional update, "later" button)
CREATE TABLE IF NOT EXISTS app_versions (
    "platform" TEXT PRIMARY KEY,
    "min_version" TEXT NOT NULL,
    "latest_version" TEXT NOT NULL,
    "store_url" TEXT,
    "updated_at" TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC') NOT NULL,

    CONSTRAINT platform_app_versions_check CHECK ("platform" IN ('android', 'ios'))
);

INSERT INTO app_versions (platform, min_version, latest_version)
VALUES ('android', '1.0.0', '1.0.0'), ('ios', '1.0.0', '1.0.0')
ON CONFLICT (platform) DO NOTHING;
