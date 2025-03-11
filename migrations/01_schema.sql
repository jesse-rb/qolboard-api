-- +goose up
CREATE TABLE IF NOT EXISTS "public"."canvases" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    "deleted_at" timestamp DEFAULT NULL,
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users",
    "canvas_data" "jsonb"
);

ALTER TABLE "public"."canvases" ENABLE ROW LEVEL SECURITY;

-- SELECT Policy
CREATE POLICY "User is canvas owner"
ON "public"."canvases"
AS PERMISSIVE
FOR SELECT
USING (
    "canvases"."user_uuid" = current_setting('myapp.user_uuid')::uuid
);

-- INSERT Policy
CREATE POLICY "User can insert canvas"
ON "public"."canvases"
AS PERMISSIVE
FOR INSERT
WITH CHECK (
    "user_uuid" = current_setting('myapp.user_uuid')::uuid
);

-- UPDATE Policy
CREATE POLICY "User can update their canvas"
ON "public"."canvases"
AS PERMISSIVE
FOR UPDATE
USING (
    "canvases"."user_uuid" = current_setting('myapp.user_uuid')::uuid
)
WITH CHECK (
    "user_uuid" = current_setting('myapp.user_uuid')::uuid
);


CREATE TABLE IF NOT EXISTS "public"."canvas_shared_invitations" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "code" "text" NOT NULL,
    "canvas_id" bigint NOT NULL REFERENCES "public"."canvases",
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users"
);


CREATE TABLE IF NOT EXISTS "public"."canvas_shared_accesses" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users",
    "canvas_id" bigint NOT NULL REFERENCES "public"."canvases",
    "canvas_shared_invitation_id" bigint NOT NULL REFERENCES "public"."canvas_shared_invitations"
);

-- +goose down
DROP TABLE IF EXISTS "public"."canvas_shared_accesses";
DROP TABLE IF EXISTS "public"."canvas_shared_invitations";
DROP TABLE IF EXISTS "public"."canvases";
