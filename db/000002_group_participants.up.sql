-- Enums?
CREATE TYPE "participant_role" AS ENUM (
  'member',
  'admin',
  'superadmin',
  -- Might be used soon
  'manager',
  'left'
);
-- Tables
CREATE TABLE IF NOT EXISTS "contact" (
  id bigserial NOT NULL,
  -- Please ensure this JID server is @lid
  jid text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at timestamp with time zone,
  push_name text,
  custom_name text,
  CONSTRAINT contact_pk PRIMARY KEY (id),
  CONSTRAINT contact_jid_unique UNIQUE (jid)
);
CREATE TABLE IF NOT EXISTS "community" (
  id bigserial NOT NULL,
  -- Please ensure this JID server is @g.us
  jid text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at timestamp with time zone,
  name text NOT NULL,
  CONSTRAINT community_pk PRIMARY KEY (id),
  CONSTRAINT community_jid_unique UNIQUE (jid)
);
CREATE TABLE IF NOT EXISTS "group" (
  id bigserial NOT NULL,
  -- Please ensure this JID server is @g.us
  jid text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at timestamp with time zone,
  name text NOT NULL,
  community_id bigint,
  CONSTRAINT group_pk PRIMARY KEY (id),
  CONSTRAINT group_jid_unique UNIQUE (jid),
  CONSTRAINT group_community_fk FOREIGN KEY (community_id) REFERENCES community (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE IF NOT EXISTS "participant" (
  id bigserial NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at timestamp with time zone,
  group_id bigint NOT NULL,
  -- The participant
  contact_id bigint NOT NULL,
  "role" participant_role NOT NULL,
  CONSTRAINT participant_pk PRIMARY KEY (id),
  CONSTRAINT participant_group_fk FOREIGN KEY (group_id) REFERENCES "group" (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT participant_contact_fk FOREIGN KEY (contact_id) REFERENCES contact (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT participant_contact_group_unique UNIQUE (group_id, contact_id)
);