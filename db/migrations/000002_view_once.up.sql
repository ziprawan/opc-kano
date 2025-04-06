BEGIN;
-- Tables
CREATE TABLE IF NOT EXISTS request_view_once (
  id bigserial NOT NULL,
  created_at timestamp with time zone DEFAULT now(),
  entity_id bigint NOT NULL,
  confirm_msg_id text NOT NULL,
  requested_msg_id text NOT NULL,
  accepted boolean,
  CONSTRAINT request_view_once_pk PRIMARY KEY (id),
  CONSTRAINT request_view_once_entity_fk FOREIGN KEY (entity_id) REFERENCES entity (id) MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE NOT VALID,
  CONSTRAINT request_view_once_entity_requested UNIQUE (entity_id, requested_msg_id)
);
-- Views
CREATE VIEW request_view_once_entity AS
SELECT rvo.id,
  rvo.created_at,
  rvo.confirm_msg_id,
  rvo.requested_msg_id,
  rvo.accepted,
  e.jid,
  e.account_id
FROM request_view_once AS rvo
  INNER JOIN entity e ON e.id = rvo.entity_id;
COMMIT;