-- Create "profiles" table
CREATE TABLE "public"."profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "display_name" character varying(255) NULL,
  "bio" text NULL,
  "headline" character varying(500) NULL,
  "locale" character varying(35) NULL,
  "timezone" character varying(64) NULL,
  "phone" character varying(40) NULL,
  "avatar_storage_key" character varying(1024) NULL,
  "links" jsonb NULL,
  "preferences" jsonb NULL,
  "status" character varying(32) NOT NULL DEFAULT 'active',
  PRIMARY KEY ("id")
);
-- Create index "idx_profiles_user_id" to table: "profiles"
CREATE UNIQUE INDEX "idx_profiles_user_id" ON "public"."profiles" ("user_id");
-- Create "profile_media" table
CREATE TABLE "public"."profile_media" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "profile_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "storage_key" character varying(1024) NOT NULL,
  "public_url" character varying(2048) NULL,
  "kind" character varying(32) NOT NULL,
  "caption" character varying(500) NULL,
  "sort_order" bigint NOT NULL DEFAULT 0,
  "width" bigint NULL,
  "height" bigint NULL,
  "content_type" character varying(128) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_profiles_media" FOREIGN KEY ("profile_id") REFERENCES "public"."profiles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_profile_media_profile_id" to table: "profile_media"
CREATE INDEX "idx_profile_media_profile_id" ON "public"."profile_media" ("profile_id");
