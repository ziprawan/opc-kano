CREATE TABLE IF NOT EXISTS "wordle" (
  id serial,
  word text NOT NULL,
  point int NOT NULL,
  lang varchar(5) NOT NULL,
  is_wordle bool NOT NULL DEFAULT false,
  CONSTRAINT wordle_pk PRIMARY KEY (id),
  CONSTRAINT wordle_word_unique UNIQUE (word, lang)
);
CREATE TABLE IF NOT EXISTS "user_wordle" (
  id serial,
  target_id int NOT NULL,
  guesses text [] NOT NULL,
  -- FK
  date_str text NOT NULL,
  user_id int NOT NULL,
  -- Constraints
  CONSTRAINT user_wordle_pk PRIMARY KEY (id),
  CONSTRAINT user_wordle_unique UNIQUE (date_str, user_id),
  CONSTRAINT user_wordle_contact_fk FOREIGN KEY (user_id) REFERENCES contact (id),
  CONSTRAINT user_wordle_wordle_fk FOREIGN KEY (target_id) REFERENCES wordle (id)
);