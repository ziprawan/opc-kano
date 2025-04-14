BEGIN;
CREATE TABLE IF NOT EXISTS contact_settings (
    id bigint NOT NULL,
    confess_target_id bigint, -- Group ID
    CONSTRAINT contact_settings_pk PRIMARY KEY (id),
    CONSTRAINT contact_settings_id_fk FOREIGN KEY (id) REFERENCES "contact" (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT contact_settings_confess_target_fk FOREIGN KEY (confess_target_id) REFERENCES "group" (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
COMMIT;