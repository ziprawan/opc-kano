BEGIN;
CREATE TABLE "wordle" (
  "id" serial NOT NULL,
  "word" text,
  "points" int NOT NULL,
  "lang" text NOT NULL,
  "length" int NOT NULL,
  CONSTRAINT "wordle_pk" PRIMARY KEY ("id")
);
COMMIT;