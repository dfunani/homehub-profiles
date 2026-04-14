-- Modify "profiles" table
ALTER TABLE "public"."profiles" ADD CONSTRAINT "uni_profiles_user_id" UNIQUE ("user_id");
