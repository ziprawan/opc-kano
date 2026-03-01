CREATE TABLE IF NOT EXISTS "sawit" (
  participant_id int NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  -- Actual data
  last_grow_date text NOT NULL,
  height int NOT NULL DEFAULT 0,
  -- Accumulative attack data
  attack_total int NOT NULL DEFAULT 0,
  attack_win int NOT NULL DEFAULT 0,
  attack_acquired_height int NOT NULL DEFAULT 0,
  attack_lost_height int NOT NULL DEFAULT 0,
  -- Constraints
  CONSTRAINT sawit_pk PRIMARY KEY (participant_id),
  CONSTRAINT sawit_participant_fk FOREIGN KEY (participant_id) REFERENCES participant (id)
);
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