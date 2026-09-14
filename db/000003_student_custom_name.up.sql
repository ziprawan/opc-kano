-- AlterTable
ALTER TABLE "students"
ADD COLUMN "custom_name" TEXT;
-- CreateIndex
CREATE INDEX "student_custom_name_idx" ON "students" USING GIST ("custom_name" gist_trgm_ops);