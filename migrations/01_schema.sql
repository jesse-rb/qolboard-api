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
-- users table
--
CREATE TABLE IF NOT EXISTS "public"."users" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "email" varchar NOT NULL,
    "email_verification_code" varchar DEFAULT NULL,
    "email_verification_code_iat" timestamp DEFAULT NULL,
    "login_otp" varchar DEFAULT NULL,
    "login_otp_iat" timestamp DEFAULT NULL,
    "verified_at" timestamp DEFAULT NULL,
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    "deleted_at" timestamp DEFAULT NULL
);

--
-- canvases table
--
CREATE TABLE IF NOT EXISTS "public"."canvases" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL,
    "deleted_at" timestamp DEFAULT NULL,
    "user_id" "uuid" NOT NULL REFERENCES "public"."users",
    "canvas_data" "jsonb"
);

-- canvas_shared_invitations table
CREATE TABLE IF NOT EXISTS "public"."canvas_shared_invitations" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "code" "text" NOT NULL,
    "canvas_id" "uuid" NOT NULL REFERENCES "public"."canvases",
    "user_id" "uuid" NOT NULL REFERENCES "public"."users"
);

-- canvas_shared_accesses table
CREATE TABLE IF NOT EXISTS "public"."canvas_shared_accesses" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_at" timestamp,
    "updated_at" timestamp,
    "deleted_at" timestamp,
    "user_id" "uuid" NOT NULL REFERENCES "public"."users",
    "canvas_id" "uuid" NOT NULL REFERENCES "public"."canvases",
    "canvas_shared_invitation_id" "uuid" NOT NULL REFERENCES "public"."canvas_shared_invitations"
);

-- +goose down

DROP TABLE IF EXISTS "public"."canvas_shared_accesses";
DROP TABLE IF EXISTS "public"."canvas_shared_invitations";
DROP TABLE IF EXISTS "public"."canvases";
DROP TABLE IF EXISTS "public"."users";

DROP FUNCTION IF EXISTS reset_user_uuid();
DROP FUNCTION IF EXISTS set_user_uuid(user_uuid text);
DROP FUNCTION IF EXISTS get_user_uuid();
