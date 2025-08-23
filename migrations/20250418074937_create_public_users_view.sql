-- +goose Up
-- +goose StatementBegin
CREATE VIEW public.view_users
AS SELECT id, email
    FROM "auth"."users";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS "public"."view_users";
-- +goose StatementEnd
