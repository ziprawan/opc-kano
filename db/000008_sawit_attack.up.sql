DROP TABLE IF EXISTS "sawit_attack";
CREATE TABLE IF NOT EXISTS "sawit_attack" (
  id serial NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  updated_at timestamp NOT NULL DEFAULT now(),
  -- The main data
  participant_id int NOT NULL,
  group_id int NOT NULL,
  message_id text NOT NULL,
  attack_size int NOT NULL,
  accepted_by int,
  is_attacker_win bool,
  -- Constraints
  CONSTRAINT sawitAttack_pk PRIMARY KEY (id),
  CONSTRAINT sawitAttack_group_fk FOREIGN KEY (group_id) REFERENCES "group" (id),
  CONSTRAINT sawitAttack_participant_fk FOREIGN KEY (participant_id) REFERENCES participant (id),
  CONSTRAINT sawitAttack_accepted_participant_fk FOREIGN KEY (accepted_by) REFERENCES participant (id),
  CONSTRAINT sawitAttack_unique UNIQUE (group_id, message_id)
)