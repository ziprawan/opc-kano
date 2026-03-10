CREATE TABLE IF NOT EXISTS "group_settings" (
  id SERIAL,
  is_game_allowed BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT group_settings_pk PRIMARY KEY (id),
  CONSTRAINT group_settings_fk FOREIGN KEY (id) REFERENCES "group" (id)
);