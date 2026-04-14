-- Drop index "idx_profile_media_profile_id" from table: "profile_media"
DROP INDEX "public"."idx_profile_media_profile_id";
-- Modify "profile_media" table
ALTER TABLE "public"."profile_media" ALTER COLUMN "kind" SET DEFAULT 'other', DROP COLUMN "sort_order", DROP COLUMN "deleted", DROP COLUMN "deleted_at";
-- Drop index "idx_profiles_user_id" from table: "profiles"
DROP INDEX "public"."idx_profiles_user_id";
-- Modify "profiles" table
ALTER TABLE "public"."profiles" ALTER COLUMN "created_at" SET DEFAULT now(), ALTER COLUMN "links" SET DEFAULT '[]', ALTER COLUMN "preferences" SET DEFAULT '{}', ALTER COLUMN "status" SET DEFAULT 'new', DROP COLUMN "deleted", DROP COLUMN "deleted_at", ADD COLUMN "status_updated_at" timestamptz NOT NULL DEFAULT now();
-- Create index "idx_profiles_updated_at" to table: "profiles"
CREATE INDEX "idx_profiles_updated_at" ON "public"."profiles" ("updated_at");
