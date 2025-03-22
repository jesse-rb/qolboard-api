-- +goose up

--
-- Useful functions
--

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_user_uuid() RETURNS uuid AS $$
    SELECT current_setting('myapp.user_uuid')::uuid;
$$ LANGUAGE SQL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_user_uuid(user_uuid text)
    RETURNS void AS $$
BEGIN
    PERFORM set_config('myapp.user_uuid', user_uuid, true);
END;
$$ LANGUAGE PLPGSQL;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION reset_user_uuid() RETURNS void AS $$
    RESET myapp.user_uuid;
$$ LANGUAGE SQL;
-- +goose StatementEnd

--
-- canvases table
--
CREATE TABLE IF NOT EXISTS "public"."canvases" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    "deleted_at" timestamp DEFAULT NULL,
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users",
    "canvas_data" "jsonb"
);

-- canvas_shared_invitations table
CREATE TABLE IF NOT EXISTS "public"."canvas_shared_invitations" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "code" "text" NOT NULL,
    "canvas_id" bigint NOT NULL REFERENCES "public"."canvases",
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users"
);

-- canvas_shared_accesses table
CREATE TABLE IF NOT EXISTS "public"."canvas_shared_accesses" (
    "id" bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "user_uuid" "uuid" NOT NULL REFERENCES "auth"."users",
    "canvas_id" bigint NOT NULL REFERENCES "public"."canvases",
    "canvas_shared_invitation_id" bigint NOT NULL REFERENCES "public"."canvas_shared_invitations"
);

--
-- canvases RLS policies
--

ALTER TABLE "public"."canvases" ENABLE ROW LEVEL SECURITY;

-- SELECT Policy
CREATE POLICY "User is canvas owner"
ON "public"."canvases"
AS PERMISSIVE
FOR SELECT
TO qolboard_api
USING (
    "canvases"."user_uuid" = get_user_uuid()
);

CREATE POLICY "User has access to canvas"
ON "public"."canvases"
AS PERMISSIVE
FOR SELECT
TO qolboard_api
USING (
    EXISTS (
        SELECT *
        FROM "public"."canvas_shared_accesses" csa
        WHERE csa.user_uuid = get_user_uuid()
        AND csa.canvas_id = "public"."canvases".id
    )
);

-- INSERT Policy
CREATE POLICY "User can insert canvas"
ON "public"."canvases"
AS PERMISSIVE
FOR INSERT
TO qolboard_api
WITH CHECK (
    "user_uuid" = get_user_uuid()
);

-- UPDATE Policy
CREATE POLICY "User can update their canvas"
ON "public"."canvases"
AS PERMISSIVE
FOR UPDATE
TO qolboard_api
USING (
    "canvases"."user_uuid" = get_user_uuid()
)
WITH CHECK (
    "user_uuid" = get_user_uuid()
);

--
-- canvas_shared_invitations policies
--

ALTER TABLE "public"."canvas_shared_invitations" ENABLE ROW LEVEL SECURITY;

-- INSERT policy
CREATE POLICY "User can create canvas shared invitations for their own canvases"
ON "public"."canvas_shared_invitations"
AS PERMISSIVE
FOR INSERT
TO qolboard_api
WITH CHECK (
    "user_uuid" = get_user_uuid()
    AND EXISTS (
        SELECT *
        FROM "public"."canvases" c
        WHERE c.id = "public"."canvas_shared_invitations".canvas_id
        AND c.user_uuid = "public"."canvas_shared_invitations".user_uuid
    )
);

-- +goose down
DROP POLICY IF EXISTS "User can create canvas shared invitations for their own canvases" ON "public"."canvas_shared_invitations";
DROP POLICY IF EXISTS "User can update their canvas" ON "public"."canvases";
DROP POLICY IF EXISTS "User has access to canvas" ON "public"."canvases";
 

DROP TABLE IF EXISTS "public"."canvas_shared_accesses";
DROP TABLE IF EXISTS "public"."canvas_shared_invitations";
DROP TABLE IF EXISTS "public"."canvases";

DROP FUNCTION IF EXISTS reset_user_uuid();
DROP FUNCTION IF EXISTS set_user_uuid(user_uuid text);
DROP FUNCTION IF EXISTS get_user_uuid();
