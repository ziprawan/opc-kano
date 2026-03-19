CREATE TABLE IF NOT EXISTS "vo_request" (
  id serial NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL DEFAULT now(),
  deleted_at timestamp,
  -- Message Info
  chat_jid text NOT NULL,
  message_id text NOT NULL,
  -- Request Info
  requester_jid text NOT NULL,
  message_owner_jid text NOT NULL,
  approval_message_id text NOT NULL,
  accepted boolean,
  -- Media info
  url text,
  direct_path text,
  media_key text NOT NULL,
  file_sha256 text NOT NULL,
  file_enc_sha256 text NOT NULL,
  media_type text NOT NULL,
  -- Constraints
  CONSTRAINT vo_request_pk PRIMARY KEY (id),
  CONSTRAINT vo_message_unique UNIQUE (chat_jid, message_id)
);