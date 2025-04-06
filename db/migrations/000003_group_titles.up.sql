BEGIN;
CREATE TABLE IF NOT EXISTS group_title (
  id bigserial NOT NULL,
  group_id bigint NOT NULL,
  title_name varchar(255) NOT NULL,
  claimable boolean NOT NULL DEFAULT true,
  CONSTRAINT group_title_pk PRIMARY KEY (id),
  CONSTRAINT group_title_group_fk FOREIGN KEY (group_id) REFERENCES "group" (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE NOT VALID,
  CONSTRAINT group_title_group_title_name_unique UNIQUE (group_id, title_name)
);
CREATE TABLE IF NOT EXISTS group_title_holder (
  id bigserial NOT NULL,
  group_title_id bigint NOT NULL,
  participant_id bigint NOT NULL,
  holding boolean NOT NULL DEFAULT true,
  CONSTRAINT group_title_holder_pk PRIMARY KEY (id),
  CONSTRAINT group_title_holder_group_title_fk FOREIGN KEY (group_title_id) REFERENCES group_title (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE NOT VALID,
  CONSTRAINT group_title_holder_participant_fk FOREIGN KEY (participant_id) REFERENCES participant (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE NOT VALID,
  CONSTRAINT group_title_holder_group_title_participant_unique UNIQUE (group_title_id, participant_id)
);
COMMIT;