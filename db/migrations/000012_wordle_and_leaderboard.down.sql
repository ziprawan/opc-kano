BEGIN;
ALTER TABLE wordle DROP COLUMN "is_wordle";
ALTER TABLE contact_settings DROP COLUMN "wordle_streaks";
ALTER TABLE contact_settings DROP COLUMN "game_points";
COMMIT;