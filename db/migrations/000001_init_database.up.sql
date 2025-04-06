BEGIN;
-- Types
CREATE TYPE chat_type AS ENUM ('GROUP', 'CONTACT');
CREATE TYPE participant_role AS ENUM ('MEMBER', 'ADMIN', 'SUPERADMIN', 'MANAGER');
-- Tables
CREATE TABLE IF NOT EXISTS account (
  id bigserial NOT NULL,
  name text NOT NULL,
  jid text NOT NULL,
  CONSTRAINT account_pk PRIMARY KEY (id),
  CONSTRAINT account_name_unique UNIQUE (name)
);
CREATE TABLE IF NOT EXISTS entity (
  id bigserial NOT NULL,
  type chat_type NOT NULL,
  jid text NOT NULL,
  account_id bigint NOT NULL,
  CONSTRAINT entity_pk PRIMARY KEY (id),
  CONSTRAINT entity_jid_account UNIQUE (jid, account_id),
  CONSTRAINT entity_account_fk FOREIGN KEY (account_id) REFERENCES account (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE NOT VALID
);
CREATE TABLE IF NOT EXISTS contact (
  id bigserial NOT NULL,
  entity_id bigint NOT NULL,
  jid text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  custom_name text,
  push_name text,
  account_id bigint NOT NULL,
  CONSTRAINT contact_pk PRIMARY KEY (id),
  CONSTRAINT contact_entity_id UNIQUE (entity_id),
  CONSTRAINT contact_account_fk FOREIGN KEY (account_id) REFERENCES account (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT contact_entity_fk FOREIGN KEY (entity_id) REFERENCES entity (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE IF NOT EXISTS "group" (
  id bigserial NOT NULL,
  account_id bigint NOT NULL,
  entity_id bigint NOT NULL,
  jid text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  owner_jid text NOT NULL,
  name text NOT NULL,
  name_set_at timestamp with time zone NOT NULL,
  name_set_by text NOT NULL,
  topic text NOT NULL,
  topic_id text NOT NULL,
  topic_set_at timestamp with time zone NOT NULL,
  topic_set_by text NOT NULL,
  topic_deleted text NOT NULL,
  is_locked boolean NOT NULL DEFAULT false,
  is_announce boolean NOT NULL DEFAULT false,
  announce_version_id text NOT NULL,
  is_ephemeral boolean NOT NULL DEFAULT false,
  disappearing_timer integer NOT NULL DEFAULT 0,
  is_incognito bool NOT NULL DEFAULT false,
  is_parent bool NOT NULL DEFAULT false,
  default_membership_approval_mode text NOT NULL,
  linked_parent_jid text NOT NULL,
  is_default_subgroup boolean NOT NULL DEFAULT false,
  is_join_approval_required boolean NOT NULL DEFAULT false,
  group_created timestamp with time zone NOT NULL,
  participant_version_id text NOT NULL,
  member_add_mode text NOT NULL DEFAULT 'all_member_add',
  CONSTRAINT group_pk PRIMARY KEY (id),
  CONSTRAINT group_entity_id UNIQUE (entity_id),
  CONSTRAINT group_account_fk FOREIGN KEY (account_id) REFERENCES account (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT group_entity_fk FOREIGN KEY (entity_id) REFERENCES entity (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE IF NOT EXISTS participant (
  id bigserial NOT NULL,
  group_id bigint NOT NULL,
  contact_id bigint NOT NULL,
  role participant_role NOT NULL,
  CONSTRAINT participant_pk PRIMARY KEY (id),
  CONSTRAINT participant_contact_fk FOREIGN KEY (contact_id) REFERENCES contact (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT participant_group_fk FOREIGN KEY (group_id) REFERENCES "group" (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT participant_contact_group UNIQUE (group_id, contact_id)
);
CREATE TABLE IF NOT EXISTS message (
  id bigserial NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now (),
  updated_at timestamp with time zone NOT NULL DEFAULT now (),
  message_id character varying(100) NOT NULL,
  entity_id bigint NOT NULL,
  "raw" text NOT NULL,
  deleted boolean DEFAULT false,
  text text,
  CONSTRAINT message_pk PRIMARY KEY (id),
  CONSTRAINT message_id_entity UNIQUE (id, entity_id),
  CONSTRAINT message_entity_fk FOREIGN KEY (entity_id) REFERENCES entity (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE
);
-- Indexes
CREATE INDEX IF NOT EXISTS "message_id_entity_id_idx" ON "message" USING "btree" ("message_id" ASC, "entity_id" ASC) WITH (DEDUPLICATE_ITEMS = TRUE);
CREATE INDEX IF NOT EXISTS "entity_jid_account_id_idx" ON "entity" USING "btree" ("jid" ASC, "account_id" ASC) WITH (DEDUPLICATE_ITEMS = TRUE);
CREATE INDEX IF NOT EXISTS "account_id_idx" ON "account" USING "btree" ("id" ASC NULLS LAST) WITH (DEDUPLICATE_ITEMS = TRUE);
-- Views
CREATE VIEW "group_participants" AS
SELECT "g"."account_id" AS "account_id",
  "p"."id" AS "participant_id",
  "e"."id" AS "entity_id",
  "g"."id" AS "group_id",
  "c"."id" AS "contact_id",
  "g"."jid" AS "group_jid",
  "c"."jid" AS "participant_jid",
  "c"."custom_name" AS "participant_custom_name",
  "c"."push_name" AS "participant_push_name"
FROM "participant" AS "p"
  INNER JOIN "group" AS "g" ON "g"."id" = "p"."group_id"
  INNER JOIN "contact" AS "c" ON "c"."id" = "p"."contact_id"
  INNER JOIN "entity" AS "e" ON "e"."id" = "g"."entity_id";
CREATE VIEW "message_with_jid" AS WITH "indexed_message" AS (
  SELECT "id",
    "message_id",
    "entity_id",
    "raw",
    "deleted",
    "text",
    "created_at",
    "updated_at"
  FROM "message"
)
SELECT "m"."id" AS "id",
  "m"."created_at",
  "m"."updated_at",
  "m"."message_id",
  "m"."raw",
  "m"."deleted",
  "m"."text",
  "e"."jid" AS "entity_jid",
  "e"."account_id" AS "account_id"
FROM "indexed_message" "m"
  JOIN "entity" "e" ON "e"."id" = "m"."entity_id";
COMMIT;