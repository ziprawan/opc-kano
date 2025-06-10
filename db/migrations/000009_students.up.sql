BEGIN;
CREATE TABLE "faculty" (
  "id" int NOT NULL,
  "name" text NOT NULL,
  CONSTRAINT "faculty_pk" PRIMARY KEY ("id")
);
CREATE TABLE "major" (
  "id" int NOT NULL,
  "name" text NOT NULL,
  "faculty_id" int NOT NULL,
  "code" text NOT NULL,
  CONSTRAINT "major_pk" PRIMARY KEY ("id"),
  CONSTRAINT "major_faculty_fk" FOREIGN KEY ("faculty_id") REFERENCES "faculty" ("id") MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE "student" (
  "id" int NOT NULL,
  "name" text NOT NULL,
  "major_id" int NOT NULL,
  "nim" int NOT NULL,
  CONSTRAINT "student_pk" PRIMARY KEY ("id"),
  CONSTRAINT "student_major_fk" FOREIGN KEY ("major_id") REFERENCES "major" ("id") MATCH SIMPLE ON UPDATE CASCADE ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED
);
COMMIT;