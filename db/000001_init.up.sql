CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE TABLE IF NOT EXISTS "students" (
  id integer NOT NULL,
  name text NOT NULL,
  nim integer NOT NULL,
  major text NOT NULL,
  faculty text NOT NULL,
  CONSTRAINT student_pk PRIMARY KEY (id) DEFERRABLE INITIALLY DEFERRED,
  CONSTRAINT student_nim UNIQUE (nim),
  CONSTRAINT students_id_key UNIQUE (id)
);
CREATE INDEX "student_name_idx" ON "students" USING GIST ("name" "gist_trgm_ops");