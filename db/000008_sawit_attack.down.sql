DROP TABLE IF EXISTS "sawit_attack";
CREATE TABLE IF NOT EXISTS "sawit_attack" (
  participant_id int NOT NULL,
  created_at timestamp NOT NULL,
  -- The attack message
  message_id text NOT NULL,
  bet_size int NOT NULL,
  accepted_by int NOT NULL,
  -- Constraints
  CONSTRAINT sawitAttack_pk PRIMARY KEY (participant_id),
  CONSTRAINT sawitAttack_participant_fk FOREIGN KEY (participant_id) REFERENCES participant (id),
  CONSTRAINT sawitAttack_accepted_participant_fk FOREIGN KEY (accepted_by) REFERENCES participant (id),
  CONSTRAINT sawitAttack_participant_message_unique UNIQUE (participant_id, message_id)
);