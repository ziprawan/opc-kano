BEGIN;
--Views
DROP VIEW IF EXISTS group_participants;
DROP VIEW IF EXISTS message_with_jid;
DROP VIEW IF EXISTS group_account;
DROP VIEW IF EXISTS contact_account;
-- Tables
DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS participant;
DROP TABLE IF EXISTS "group";
DROP TABLE IF EXISTS contact;
DROP TABLE IF EXISTS entity;
DROP TABLE IF EXISTS account;
-- Types
DROP TYPE chat_type;
DROP TYPE participant_role;
COMMIT;