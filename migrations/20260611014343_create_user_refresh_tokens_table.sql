-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "public"."user_refresh_tokens"(
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" "uuid" NOT NULL REFERENCES "public"."users",
    "refresh_token" VARCHAR NOT NULL UNIQUE,
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    "deleted_at" timestamp DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "public"."user_refresh_tokens";
-- +goose StatementEnd
