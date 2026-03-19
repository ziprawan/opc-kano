ALTER TABLE IF EXISTS "contact" DROP COLUMN IF EXISTS "confess_target";
ALTER TABLE IF EXISTS "group" DROP COLUMN IF EXISTS "is_announcement";
ALTER TABLE IF EXISTS "group_settings" DROP COLUMN IF EXISTS "is_confess_allowed";