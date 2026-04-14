-- Modify "profile_media" table
ALTER TABLE "public"."profile_media" ADD COLUMN "deleted" boolean NOT NULL DEFAULT false, ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "idx_profile_media_deleted_at" to table: "profile_media"
CREATE INDEX "idx_profile_media_deleted_at" ON "public"."profile_media" ("deleted_at");
-- Modify "profiles" table
ALTER TABLE "public"."profiles" ADD COLUMN "deleted" boolean NOT NULL DEFAULT false, ADD COLUMN "deleted_at" timestamptz NULL;
-- Create index "idx_profiles_deleted_at" to table: "profiles"
CREATE INDEX "idx_profiles_deleted_at" ON "public"."profiles" ("deleted_at");
