-- CreateExtension
CREATE EXTENSION IF NOT EXISTS "pg_ivm";
-- CreateEnum
CREATE TYPE "subject_category" AS ENUM ('LECTURE', 'RESEARCH', 'INTERNSHIP', 'THESIS');
-- CreateEnum
CREATE TYPE "schedule_activity" AS ENUM (
  'LECTURE',
  'TUTORIAL',
  'LAB_WORK',
  'QUIZ',
  'MIDTERM',
  'FINAL'
);
-- CreateEnum
CREATE TYPE "schedule_method" AS ENUM ('IN_PERSON', 'ONLINE', 'HYBRID');
-- CreateEnum
CREATE TYPE "strata" AS ENUM ('S1', 'S2', 'S3', 'PR');
-- CreateEnum
CREATE TYPE "campus" AS ENUM ('JATINANGOR', 'GANESHA', 'CIREBON', 'JAKARTA');
-- CreateTable
CREATE TABLE "major" (
  "id" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "faculty" TEXT NOT NULL,
  CONSTRAINT "major_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "curricula" (
  "year" INTEGER NOT NULL,
  CONSTRAINT "curricula_pk" PRIMARY KEY ("year")
);
-- CreateTable
CREATE TABLE "semester" (
  "id" SERIAL NOT NULL,
  "year" INTEGER NOT NULL,
  "semester" INTEGER NOT NULL,
  "end" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "start" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "semester_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "subject" (
  "id" INTEGER NOT NULL,
  "code" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "sks" INTEGER NOT NULL,
  "category" "subject_category",
  "curricula_year" INTEGER NOT NULL,
  CONSTRAINT "subject_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "lecturer" (
  "id" SERIAL NOT NULL,
  "name" TEXT NOT NULL,
  CONSTRAINT "lecturer_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "lecturer_in_class" (
  "lecturer_id" INTEGER NOT NULL,
  "subject_class_id" INTEGER NOT NULL,
  CONSTRAINT "lecturer_in_class_pk" PRIMARY KEY ("lecturer_id", "subject_class_id")
);
-- CreateTable
CREATE TABLE "subject_class" (
  "id" INTEGER NOT NULL,
  "number" INTEGER NOT NULL,
  "edunex_class_id" INTEGER,
  "teams_link" TEXT,
  "major_id" INTEGER NOT NULL,
  "quota" INTEGER,
  "semester_id" INTEGER NOT NULL,
  "subject_id" INTEGER NOT NULL,
  CONSTRAINT "subject_class_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "class_constraint" (
  "id" SERIAL NOT NULL,
  "other" TEXT [],
  "faculties" TEXT [],
  "stratas" "strata" [],
  "campuses" "campus" [],
  "subject_class_id" INTEGER NOT NULL,
  "semester" INTEGER [],
  CONSTRAINT "class_constraint_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "constraint_major" (
  "id" SERIAL NOT NULL,
  "addon_data" TEXT NOT NULL,
  "major_id" INTEGER,
  CONSTRAINT "constraint_major_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "major_in_constraint" (
  "class_constraint_id" INTEGER NOT NULL,
  "constraint_major_id" INTEGER NOT NULL,
  CONSTRAINT "major_in_constraint_pk" PRIMARY KEY ("class_constraint_id", "constraint_major_id")
);
-- CreateTable
CREATE TABLE "room" (
  "id" SERIAL NOT NULL,
  "name" TEXT NOT NULL,
  CONSTRAINT "room_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "class_schedule" (
  "id" SERIAL NOT NULL,
  "start" TIMESTAMPTZ(6) NOT NULL,
  "end" TIMESTAMPTZ(6) NOT NULL,
  "activity" "schedule_activity" NOT NULL,
  "method" "schedule_method" NOT NULL,
  "unix_start" BIGINT NOT NULL,
  "unix_end" BIGINT NOT NULL,
  "subject_class_id" INTEGER NOT NULL,
  CONSTRAINT "class_schedule_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "room_in_class" (
  "room_id" INTEGER NOT NULL,
  "class_schedule_id" INTEGER NOT NULL,
  CONSTRAINT "room_in_class_pk" PRIMARY KEY ("room_id", "class_schedule_id")
);
-- CreateTable
CREATE TABLE "edunex_class_id" (
  "id" SERIAL NOT NULL,
  "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" TIMESTAMP(3) NOT NULL,
  "sems_ctx" TEXT NOT NULL,
  "subject_code" TEXT NOT NULL,
  "class_num" INTEGER NOT NULL,
  "class_id" INTEGER,
  CONSTRAINT "edunex_class_id_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "class_reminder" (
  "id" SERIAL NOT NULL,
  "jid" TEXT NOT NULL,
  "subject_class_id" INTEGER NOT NULL,
  "anchor_at_end" BOOLEAN NOT NULL DEFAULT false,
  "offset_minutes" INTEGER NOT NULL DEFAULT 0,
  CONSTRAINT "class_reminder_pk" PRIMARY KEY ("id")
);
-- CreateTable
CREATE TABLE "class_follower" (
  "id" SERIAL NOT NULL,
  "jid" TEXT NOT NULL,
  "subject_class_id" INTEGER NOT NULL,
  CONSTRAINT "class_follower_pk" PRIMARY KEY ("id")
);
-- CreateTable
SELECT pgivm.create_immv(
    'class_reminder_view',
    'SELECT cs.id AS schedule_id, cs.subject_class_id AS class_id, cr.jid AS jid, cr.anchor_at_end AS anchor_at_end, CASE WHEN cr.anchor_at_end THEN cs.unix_end + (cr.offset_minutes * 60) ELSE cs.unix_start + (cr.offset_minutes * 60) END AS alert_time_unix FROM class_reminder cr INNER JOIN class_schedule cs ON cs.subject_class_id = cr.subject_class_id'
  );
-- CreateTable
CREATE TABLE "class_reminder_delivery" (
  "schedule_id" INTEGER NOT NULL,
  "jid" TEXT NOT NULL,
  "delivered_for_unix" BIGINT NOT NULL,
  "delivered_at" BIGINT,
  CONSTRAINT "class_reminder_delivery_pk" PRIMARY KEY ("schedule_id", "jid", "delivered_for_unix")
);
-- CreateIndex
CREATE UNIQUE INDEX "semester_year_semester_unique" ON "semester"("year", "semester");
-- CreateIndex
CREATE INDEX "subject_code_idx" ON "subject"("code");
-- CreateIndex
CREATE UNIQUE INDEX "subject_code_curricula_unique" ON "subject"("code", "curricula_year");
-- CreateIndex
CREATE UNIQUE INDEX "lecturer_name_unique" ON "lecturer"("name");
-- CreateIndex
CREATE UNIQUE INDEX "classConstraint_subjectClassId_unique" ON "class_constraint"("subject_class_id");
-- CreateIndex
CREATE UNIQUE INDEX "constraintMajor_addonData_majorId_unique" ON "constraint_major"("addon_data", "major_id");
-- CreateIndex
CREATE UNIQUE INDEX "room_name_unique" ON "room"("name");
-- CreateIndex
CREATE INDEX "classschedule_start_idx" ON "class_schedule"("start");
-- CreateIndex
CREATE INDEX "classschedule_end_idx" ON "class_schedule"("end");
-- CreateIndex
CREATE INDEX "classschedule_unix_start_idx" ON "class_schedule"("unix_start");
-- CreateIndex
CREATE INDEX "classschedule_unix_end_idx" ON "class_schedule"("unix_end");
-- CreateIndex
CREATE UNIQUE INDEX "edunexclassid_semsCtx_subjectCode_classNum_unique" ON "edunex_class_id"("sems_ctx", "subject_code", "class_num");
-- CreateIndex
CREATE INDEX "classReminder_subjectClassId_idx" ON "class_reminder"("subject_class_id");
-- CreateIndex
CREATE UNIQUE INDEX "classReminder_jid_subjectClassId_offset_unique" ON "class_reminder"(
  "jid",
  "subject_class_id",
  "offset_minutes",
  "anchor_at_end"
);
-- CreateIndex
CREATE INDEX "classFollower_subjectClassId_idx" ON "class_follower"("subject_class_id");
-- CreateIndex
CREATE UNIQUE INDEX "classFollower_jid_subjectClassId_unique" ON "class_follower"("jid", "subject_class_id");
-- CreateIndex
CREATE INDEX "student_custom_name_idx" ON "students" USING GIST ("custom_name" gist_trgm_ops);
-- AddForeignKey
ALTER TABLE "subject"
ADD CONSTRAINT "subject_curricula_fk" FOREIGN KEY ("curricula_year") REFERENCES "curricula"("year") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "lecturer_in_class"
ADD CONSTRAINT "lecturer_subjectClass_fk" FOREIGN KEY ("subject_class_id") REFERENCES "subject_class"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "lecturer_in_class"
ADD CONSTRAINT "subjectClass_lecturer_fk" FOREIGN KEY ("lecturer_id") REFERENCES "lecturer"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "subject_class"
ADD CONSTRAINT "subjectClass_major_fk" FOREIGN KEY ("major_id") REFERENCES "major"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "subject_class"
ADD CONSTRAINT "subjectClass_semester_fk" FOREIGN KEY ("semester_id") REFERENCES "semester"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "subject_class"
ADD CONSTRAINT "subjectClass_subject_fk" FOREIGN KEY ("subject_id") REFERENCES "subject"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "class_constraint"
ADD CONSTRAINT "classConstraint_subjectClass_fk" FOREIGN KEY ("subject_class_id") REFERENCES "subject_class"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "constraint_major"
ADD CONSTRAINT "constraintMajor_classConstraint_fk" FOREIGN KEY ("major_id") REFERENCES "major"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "major_in_constraint"
ADD CONSTRAINT "classConstraint_constraintMajor_fk" FOREIGN KEY ("constraint_major_id") REFERENCES "constraint_major"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "major_in_constraint"
ADD CONSTRAINT "constraintMajor_classConstraint_fk" FOREIGN KEY ("class_constraint_id") REFERENCES "class_constraint"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "class_schedule"
ADD CONSTRAINT "classSchedule_subjectClass_fk" FOREIGN KEY ("subject_class_id") REFERENCES "subject_class"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "room_in_class"
ADD CONSTRAINT "classSchedule_room_fk" FOREIGN KEY ("room_id") REFERENCES "room"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "room_in_class"
ADD CONSTRAINT "room_classSchedule_fk" FOREIGN KEY ("class_schedule_id") REFERENCES "class_schedule"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "class_reminder"
ADD CONSTRAINT "classReminder_subjectClass_fk" FOREIGN KEY ("subject_class_id") REFERENCES "subject_class"("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- AddForeignKey
ALTER TABLE "class_follower"
ADD CONSTRAINT "classFollower_subjectClass_fk" FOREIGN KEY ("subject_class_id") REFERENCES "subject_class"("id") ON DELETE CASCADE ON UPDATE CASCADE;